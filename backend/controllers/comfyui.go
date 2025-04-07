package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"VirtualHumanStudio/backend/config"
)

// ComfyUIUploadImageResponse ComfyUI上传图片响应
type ComfyUIUploadImageResponse struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Subfolder string `json:"subfolder"`
}

// ComfyUIPromptResponse ComfyUI提交任务响应
type ComfyUIPromptResponse struct {
	PromptID string `json:"prompt_id"`
}

// ComfyUIHistoryResponse ComfyUI查询任务进度响应
type ComfyUIHistoryResponse struct {
	Status string                 `json:"status"` // pending, processing, completed, failed
	Data   map[string]interface{} `json:"data"`
}

// UploadMaskToComfyUI 上传蒙版图片到ComfyUI服务器
func UploadMaskToComfyUI(maskPath string, modelRef string) (string, error) {
	fmt.Println("====", maskPath, modelRef)
	file, err := os.Open(maskPath)
	if err != nil {
		return "", fmt.Errorf("打开蒙版文件失败: %v", err)
	}
	defer file.Close()

	// 创建multipart/form-data请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加蒙版文件
	part, err := writer.CreateFormFile("image", filepath.Base(maskPath))
	if err != nil {
		return "", fmt.Errorf("创建form-data请求失败: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("复制蒙版数据失败: %v", err)
	}

	// 添加关联的模型引用
	original_ref := map[string]string{
		"type":        "input",
		"filename":    modelRef,
		"resize_mode": "resize",
	}
	original_ref_json, _ := json.Marshal(original_ref)
	fmt.Println("====", string(original_ref_json))
	_ = writer.WriteField("original_ref", string(original_ref_json))
	writer.WriteField("type", "input")
	writer.WriteField("subfolder", "clipspace")

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("关闭writer失败: %v", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s%s", config.AppConfig.ComfyUIConf.ServerURL, config.AppConfig.ComfyUIConf.UploadMaskAPI)
	resp, err := http.Post(url, writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("发送蒙版请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("上传蒙版失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var response ComfyUIUploadImageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("解析蒙版响应失败: %v", err)
	}

	return response.Name, nil
}

// UploadImageToComfyUI 上传图片到ComfyUI服务器
func UploadImageToComfyUI(imagePath string) (string, error) {
	// 打开图片文件
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("打开图片文件失败: %v", err)
	}
	defer file.Close()

	// 创建multipart/form-data请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加图片文件
	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return "", fmt.Errorf("创建form-data请求失败: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("复制图片数据失败: %v", err)
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("关闭writer失败: %v", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s%s", config.AppConfig.ComfyUIConf.ServerURL, config.AppConfig.ComfyUIConf.UploadImageAPI)
	resp, err := http.Post(url, writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("发送图片请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("上传图片失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var response ComfyUIUploadImageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("解析图片响应失败: %v", err)
	}

	return response.Name, nil
}

// BuildAccessoryPrompt 构建物品替换工作流
func BuildAccessoryPrompt(itemImageRef, modelImageRef, maskImageRef string) string {
	// 从模板文件加载工作流配置
	templatePath := filepath.Join(config.AppConfig.DataDir, "workflow/accessory.json")
	fileContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("读取工作流模板失败: %v\n", err)
		return ""
	}

	var workflow map[string]interface{}
	err = json.Unmarshal(fileContent, &workflow)
	if err != nil {
		fmt.Printf("解析工作流模板失败: %v\n", err)
		return ""
	}

	templatePath = filepath.Join(config.AppConfig.DataDir, "workflow/accessory_api.json")
	fileContent, err = os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("读取工作流模板失败: %v\n", err)
		return ""
	}
	var prompt map[string]interface{}
	err = json.Unmarshal(fileContent, &prompt)
	if err != nil {
		fmt.Printf("解析工作流模板失败: %v\n", err)
		return ""
	}

	maskImageRef = "clipspace/" + maskImageRef + " [input]"

	// 替换模板变量
	replaceVariables := map[string]string{
		"{{modelImageRef}}": modelImageRef,
		"{{maskImageRef}}":  maskImageRef,
		"{{itemImageRef}}":  itemImageRef,
		"{{defaultModel}}":  config.AppConfig.ComfyUIConf.DefaultModel,
	}

	updatedWorkflow, _ := json.Marshal(workflow)
	workflowStr := string(updatedWorkflow)

	updatedPrompt, _ := json.Marshal(prompt)
	promptStr := string(updatedPrompt)
	for k, v := range replaceVariables {
		workflowStr = strings.ReplaceAll(workflowStr, k, v)
		promptStr = strings.ReplaceAll(promptStr, k, v)
	}

	err = json.Unmarshal([]byte(workflowStr), &workflow)
	if err != nil {
		fmt.Printf("变量替换后解析工作流失败: %v\n", err)
		return ""
	}
	err = json.Unmarshal([]byte(promptStr), &prompt)
	if err != nil {
		fmt.Printf("变量替换后解析工作流失败: %v\n", err)
		return ""
	}

	// 从workflow中提取prompt

	// 创建extra_data字段，包含workflow信息
	extraData := map[string]interface{}{
		"extra_pnginfo": map[string]interface{}{
			"workflow": workflow,
		},
	}

	// 转换为最终JSON字符串
	promptJSON, err := json.Marshal(map[string]interface{}{
		"prompt":     prompt,
		"client_id":  config.AppConfig.ComfyUIConf.ClientID,
		"extra_data": extraData,
	})
	if err != nil {
		fmt.Printf("构建工作流JSON失败: %v\n", err)
		return ""
	}

	return string(promptJSON)
}

// SubmitPromptToComfyUI 提交任务到ComfyUI服务器
func SubmitPromptToComfyUI(prompt string) (string, error) {
	// 创建请求体
	body := bytes.NewBuffer([]byte(prompt))

	// 发送请求
	url := fmt.Sprintf("%s%s", config.AppConfig.ComfyUIConf.ServerURL, config.AppConfig.ComfyUIConf.PromptAPI)
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("提交任务失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var response ComfyUIPromptResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 返回任务ID
	return response.PromptID, nil
}

// QueryComfyUITaskStatus 查询任务进度
func QueryComfyUITaskStatus(promptID string) (string, map[string]interface{}, error) {
	// 发送请求
	url := fmt.Sprintf("%s%s/%s", config.AppConfig.ComfyUIConf.ServerURL, config.AppConfig.ComfyUIConf.HistoryAPI, promptID)
	fmt.Println("#####", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("查询任务进度失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应是否为空
	if response == nil || len(response) == 0 {
		return "failed", nil, fmt.Errorf("响应数据为空")
	}

	// 检查任务状态
	status := "processing"

	// 尝试获取任务数据
	data, ok := response[promptID].(map[string]interface{})
	if !ok {
		// 记录响应内容以便调试
		respBytes, _ := json.Marshal(response)
		fmt.Printf("响应数据结构异常: %s\n", string(respBytes))

		// 检查是否有其他可用的数据
		if len(response) > 0 {
			// 尝试使用第一个可用的键
			for key, value := range response {
				if mapData, mapOk := value.(map[string]interface{}); mapOk {
					fmt.Printf("使用替代键: %s\n", key)
					data = mapData
					ok = true
					break
				}
			}
		}

		// 如果仍然没有找到有效数据
		if !ok {
			return "failed", nil, fmt.Errorf("任务数据格式错误: 响应中不包含有效的任务数据")
		}
	}

	// 检查是否有输出节点的数据
	outputs, ok := data["outputs"].(map[string]interface{})
	if ok && len(outputs) > 0 {
		// 检查SaveImage节点的输出
		if nodeOutput, ok := outputs["9"].(map[string]interface{}); ok {
			if images, ok := nodeOutput["images"].([]interface{}); ok && len(images) > 0 {
				status = "completed"
			}
		} else {
			// 尝试查找任何包含images的节点
			for nodeID, nodeData := range outputs {
				if nodeMap, ok := nodeData.(map[string]interface{}); ok {
					if images, ok := nodeMap["images"].([]interface{}); ok && len(images) > 0 {
						fmt.Printf("找到替代输出节点: %s\n", nodeID)
						status = "completed"
						break
					}
				}
			}
		}
	} else if execInfo, ok := data["exec_info"].(map[string]interface{}); ok {
		// 检查执行信息
		if execStatus, ok := execInfo["status"].(string); ok {
			if execStatus == "error" || execStatus == "failed" {
				status = "failed"
			} else if execStatus == "success" || execStatus == "completed" {
				status = "completed"
			}
		}
	}

	return status, data, nil
}

// DownloadComfyUIResult 下载ComfyUI任务结果图片
func DownloadComfyUIResult(data map[string]interface{}, savePath string, nodeID string) error {
	// 从输出数据中获取结果图片信息
	outputs, ok := data["outputs"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("输出数据格式错误")
	}

	// 首先尝试获取SaveImage节点的输出（节点ID为"9"）
	nodeOutput, ok := outputs[nodeID].(map[string]interface{})

	// 如果节点ID为"9"的输出不存在，则尝试查找任何包含images的节点
	if !ok {
		fmt.Println("未找到节点ID为9的SaveImage节点，尝试查找其他输出节点...")

		// 遍历所有输出节点，查找包含images的节点
		var foundNode bool
		for nodeID, nodeData := range outputs {
			if nodeMap, nodeOk := nodeData.(map[string]interface{}); nodeOk {
				if images, imagesOk := nodeMap["images"].([]interface{}); imagesOk && len(images) > 0 {
					fmt.Printf("找到替代输出节点: %s\n", nodeID)
					nodeOutput = nodeMap
					ok = true
					foundNode = true
					break
				}
			}
		}

		// 如果仍然没有找到包含images的节点
		if !foundNode {
			// 记录所有可用的节点ID，以便调试
			nodeIDs := make([]string, 0, len(outputs))
			for nodeID := range outputs {
				nodeIDs = append(nodeIDs, nodeID)
			}
			fmt.Printf("未找到包含images的输出节点，可用节点ID: %v\n", nodeIDs)
			return fmt.Errorf("未找到包含图片的输出节点")
		}
	}

	// 获取图片信息
	images, ok := nodeOutput["images"].([]interface{})
	if !ok || len(images) == 0 {
		return fmt.Errorf("结果图片不存在")
	}

	// 获取第一张图片信息
	imageInfo, ok := images[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("图片信息格式错误")
	}

	// 获取图片文件名
	filename, ok := imageInfo["filename"].(string)
	if !ok {
		return fmt.Errorf("图片文件名不存在")
	}

	// 获取图片子文件夹，如果不存在则使用默认值
	subfolder := ""
	if subfolder, subfolderOk := imageInfo["subfolder"].(string); subfolderOk {
		subfolder = subfolder
	}

	// 获取图片URL
	imageURL := fmt.Sprintf("%s/view?filename=%s&subfolder=%s&type=temp",
		config.AppConfig.ComfyUIConf.ServerURL,
		filename,
		subfolder)

	fmt.Printf("正在下载图片: %s\n", imageURL)

	// 下载图片
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入图片数据
	bytes, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("写入图片数据失败: %v", err)
	}

	fmt.Printf("成功下载图片到: %s (大小: %d 字节)\n", savePath, bytes)
	return nil
}
