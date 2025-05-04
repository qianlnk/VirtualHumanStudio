package controllers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
)

// processImageTask 异步处理图像任务
func processImageTask(taskID uint) {
	// 获取任务信息
	var task models.ComfyUIWorkflowTask
	if err := db.DB.First(&task, taskID).Error; err != nil {
		fmt.Printf("获取任务失败: %v\n", err)
		return
	}

	// 更新任务状态为处理中
	task.Status = "processing"
	db.DB.Save(&task)

	// 解析工作流配置
	workflowPath := filepath.Join(config.AppConfig.DataDir, "workflow", task.WorkflowName+".json")
	workflowContent, err := os.ReadFile(workflowPath)
	if err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("工作流配置加载失败: %v", err)
		db.DB.Save(&task)
		return
	}
	workflowStr := string(workflowContent)

	// 解析prompt
	promptPath := filepath.Join(config.AppConfig.DataDir, "workflow", task.WorkflowName+"_api.json")
	promptContent, err := os.ReadFile(promptPath)
	if err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("prompt配置加载失败: %v", err)
		db.DB.Save(&task)
		return
	}
	promptStr := string(promptContent)

	// 解析输入参数
	var inputParams []models.InputParam
	if err := json.Unmarshal([]byte(task.InputParams), &inputParams); err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("解析输入参数失败: %v", err)
		db.DB.Save(&task)
		return
	}

	rand.NewSource(time.Now().UnixNano())
	replaceVariables := map[string]string{
		"{{seed}}": strconv.FormatInt(rand.Int63n(1000000000000), 10),
	}

	// 替换工作流中的文件路径
	for _, param := range inputParams {
		// 上传文件到ComfyUI服务器
		if param.Type == models.ParamTypeImage {
			fileRef, err := UploadImageToComfyUI(filepath.Join(config.AppConfig.DataDir, param.Value))
			if err != nil {
				task.Status = "failed"
				task.ErrorMsg = fmt.Sprintf("文件上传失败(%s): %v", param.Key, err)
				db.DB.Save(&task)
				return
			}

			replaceVariables["{{"+param.Key+"}}"] = fileRef

		} else if param.Type == models.ParamTypeMask {
			// 遮照图对应的图片xxx  遮照图xxxMask
			fileRef, err := UploadMaskToComfyUI(filepath.Join(config.AppConfig.DataDir, param.Value), replaceVariables["{{"+strings.TrimSuffix(param.Key, "Mask")+"}}"])
			if err != nil {
				task.Status = "failed"
				task.ErrorMsg = fmt.Sprintf("文件上传失败(%s): %v", param.Key, err)
				db.DB.Save(&task)
				return
			}

			replaceVariables["{{"+param.Key+"}}"] = "clipspace/" + fileRef + " [input]"

		} else {
			// 替换工作流中的文本参数
			replaceVariables["{{"+param.Key+"}}"] = param.Value
		}
	}

	for k, v := range replaceVariables {
		workflowStr = strings.ReplaceAll(workflowStr, k, v)
		promptStr = strings.ReplaceAll(promptStr, k, v)
	}

	workflow := map[string]interface{}{}
	prompt := map[string]interface{}{}
	err = json.Unmarshal([]byte(workflowStr), &workflow)
	if err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("解析工作流配置失败: %v", err)
		db.DB.Save(&task)
		return
	}

	err = json.Unmarshal([]byte(promptStr), &prompt)
	if err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("解析prompt配置失败: %v", err)
		db.DB.Save(&task)
		return
	}

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
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("构建工作流JSON失败: %v", err)
		db.DB.Save(&task)
		return
	}

	fmt.Println("=====\n", string(promptJSON))

	// 提交任务到ComfyUI服务器
	promptID, err := SubmitPromptToComfyUI(string(promptJSON))
	if err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("提交任务失败: %v", err)
		db.DB.Save(&task)
		return
	}

	// 更新任务ID
	task.TaskID = promptID
	db.DB.Save(&task)

	// 等待任务完成
	for {
		// 查询任务状态
		status, data, err := QueryComfyUITaskStatus(promptID)
		if err != nil {
			fmt.Printf("查询任务进度失败: %v\n", err)

			if status == "failed" {
				task.Status = "failed"
				task.ErrorMsg = "任务执行失败:" + err.Error()
				db.DB.Save(&task)
				return
			}

			time.Sleep(5 * time.Second)
			continue
		}

		// 更新任务状态
		task.Status = status

		// 如果任务已完成，保存输出参数
		if status == "completed" && data != nil {
			// 下载结果图片
			moduleConfig, err := getMuduleConfigByID(task.TaskType)
			if err != nil {
				task.Status = "failed"
				task.ErrorMsg = fmt.Sprintf("获取模块配置失败: %v", err)
				db.DB.Save(&task)
				return
			}

			var outputParams []models.InputParam
			for _, param := range moduleConfig.OutputParams {
				ext := ".png"
				if param.Type == models.ParamTypeImage {
					ext = ".png"
				} else if param.Type == models.ParamTypeVideo {
					ext = ".mp4"
				}

				uniqueID := uuid.New().String()
				fileName := uniqueID + ext
				filePath := utils.GetUserFilePath(task.UserID, config.AppConfig.UploadDir, fileName)
				fullFilePath := filepath.Join(config.AppConfig.DataDir, filePath)

				err := DownloadComfyUIResult(data, ext, fullFilePath, param.Key)
				if err != nil {
					task.Status = "failed"
					task.ErrorMsg = fmt.Sprintf("下载结果图片失败: %v", err)
					db.DB.Save(&task)
					return
				}

				outputParams = append(outputParams, models.InputParam{
					Key:   param.Key,
					Type:  param.Type,
					Value: filePath,
					Alias: param.Alias,
				})
			}

			// 将输出数据转换为JSON字符串
			task.OutputParams = utils.ToJSONString(outputParams)
			db.DB.Save(&task)
			break
		} else if status == "failed" {
			task.Status = "failed"
			task.ErrorMsg = "任务执行失败"
			db.DB.Save(&task)
			break
		}

		// 等待一段时间后再次查询
		time.Sleep(2 * time.Second)
	}
}

// replaceFilePathInWorkflow 替换工作流配置中的文件路径
func replaceFilePathInWorkflow(workflow map[string]interface{}, key string, filePath string) map[string]interface{} {
	// 遍历所有节点
	for _, nodeInterface := range workflow {
		node, ok := nodeInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// 检查节点类型
		classType, ok := node["class_type"].(string)
		if !ok {
			continue
		}

		// 处理LoadImage节点
		if classType == "LoadImage" {
			inputs, ok := node["inputs"].(map[string]interface{})
			if !ok {
				continue
			}

			// 检查image字段
			image, ok := inputs["image"].(string)
			if !ok {
				continue
			}

			// 检查是否包含占位符
			placeholder := fmt.Sprintf("{{%sRef}}", key)
			if image == placeholder {
				inputs["image"] = filePath
			}
		}
	}

	return workflow
}
