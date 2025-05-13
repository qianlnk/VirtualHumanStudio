package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
)

// InspirationParam 参数结构
type InspirationParam struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// InspirationTaskResponse 共享任务响应
type InspirationTaskResponse struct {
	ID           uint               `json:"id"`
	Type         string             `json:"type"`          // 任务类型: "comfyui" 或 "digital_human"
	TaskType     string             `json:"task_type"`     // 具体任务类型，如"accessory"、"workflow"等
	Name         string             `json:"name"`          // 任务名称
	Description  string             `json:"description"`   // 任务描述
	ResultURL    string             `json:"result_url"`    // 结果图片/视频URL
	InputParams  []InspirationParam `json:"input_params"`  // 输入参数
	OutputParams []InspirationParam `json:"output_params"` // 输出参数
	UserID       uint               `json:"user_id"`       // 创建者ID
	Username     string             `json:"username"`      // 创建者用户名
	CreatedAt    string             `json:"created_at"`    // 创建时间
}

// GetInspirationTasks 获取灵感页面任务
func GetInspirationTasks(c *gin.Context) {
	// 分页参数
	page, size := utils.GetPaginationParams(c)

	shareTasks := []*models.ShareTask{}

	var responses []InspirationTaskResponse
	var totalCount int64

	// 查询总数
	db.DB.Model(&models.ShareTask{}).
		Where("share_status = ?", "approved").
		Count(&totalCount)

	db.DB.Where("share_status = ?", "approved").
		Order("id DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&shareTasks)

	for _, task := range shareTasks {
		fmt.Println("task:", utils.ToJSONString(task))
		if task.Mode == "comfyui" {
			var comfyUITask models.ComfyUIWorkflowTask
			db.DB.Where("id = ?", task.TaskID).First(&comfyUITask)

			responses = append(responses, convertComfyUITaskToResponse(comfyUITask))

		} else if task.Mode == "digital_human" {
			var digitalHuman models.DigitalHuman
			db.DB.Where("id = ?", task.TaskID).First(&digitalHuman)
			responses = append(responses, convertDigitalHumanToResponse(digitalHuman))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": responses,
		"total": totalCount,
		"page":  page,
		"size":  size,
	})
}

// GetInspirationDetail 获取单个灵感详情
func GetInspirationDetail(c *gin.Context) {
	id := c.Param("id")
	taskID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	// 尝试查找ComfyUI任务
	var comfyUITask models.ComfyUIWorkflowTask
	if err := db.DB.Where("id = ? AND share_status = ?", taskID, "approved").First(&comfyUITask).Error; err == nil {
		// 找到了ComfyUI任务，转换为响应格式
		response := convertComfyUITaskToResponse(comfyUITask)
		c.JSON(http.StatusOK, response)
		return
	}

	// 尝试查找数字人任务
	var digitalHuman models.DigitalHuman
	if err := db.DB.Where("id = ? AND share_status = ?", taskID, "approved").First(&digitalHuman).Error; err == nil {
		// 找到了数字人任务，转换为响应格式
		response := convertDigitalHumanToResponse(digitalHuman)
		c.JSON(http.StatusOK, response)
		return
	}

	// 未找到任务
	c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定任务"})
}

// 从ComfyUI工作流任务转换为响应结构体
func convertComfyUITaskToResponse(task models.ComfyUIWorkflowTask) InspirationTaskResponse {
	var resultURL string
	var inputParams, outputParams []InspirationParam

	if task.TaskType == "accessory" {
		var accessory models.Accessory
		db.DB.Where("id = ?", task.ID).First(&accessory)
		resultURL = utils.GetFileURL(accessory.ResultImage)

		// 设置输入参数
		inputParams = []InspirationParam{
			{Key: "item_image", Label: "物品图", Type: "image", Value: utils.GetFileURL(accessory.ItemImage)},
			{Key: "model_image", Label: "模特图", Type: "image", Value: utils.GetFileURL(accessory.ModelImage)},
			{Key: "mask_image", Label: "蒙版图", Type: "mask", Value: utils.GetFileURL(accessory.MaskImage)},
		}

		// 设置输出参数
		outputParams = []InspirationParam{
			{Key: "result_image", Label: "结果图", Type: "image", Value: utils.GetFileURL(accessory.ResultImage)},
		}
	} else {
		// 查询工作流任务获取更多信息
		var workflowTask models.ComfyUIWorkflowTask
		if db.DB.Where("id = ?", task.ID).First(&workflowTask).Error == nil {
			// 解析输入参数
			if workflowTask.InputParams != "" {
				var params []models.InputParam
				json.Unmarshal([]byte(workflowTask.InputParams), &params)

				for _, param := range params {
					inputParam := InspirationParam{
						Key:   param.Key,
						Label: param.Alias,
						Type:  param.Type,
						Value: param.Value,
					}

					// 对于图片类型参数，生成完整URL
					if param.Type == "image" || param.Type == "mask" || param.Type == "video" || param.Type == "audio" {
						inputParam.Value = utils.GetFileURL(param.Value)
					}

					inputParams = append(inputParams, inputParam)
				}
			}

			// 解析输出参数获取结果图片
			if workflowTask.OutputParams != "" {
				json.Unmarshal([]byte(workflowTask.OutputParams), &outputParams)

				// 设置结果URL为第一个图片类型的输出参数
				for i, param := range outputParams {
					if param.Type == "image" || param.Type == "mask" || param.Type == "video" || param.Type == "audio" {
						resultURL = utils.GetFileURL(param.Value)
						outputParams[i].Value = utils.GetFileURL(param.Value)
					}
				}
			}
		}
	}

	// 获取用户名
	username := "用户"
	if task.UserID > 0 {
		var user models.User
		db.DB.Select("username").Where("id = ?", task.UserID).First(&user)
		if user.Username != "" {
			username = user.Username
		}
	}

	return InspirationTaskResponse{
		ID:           task.ID,
		Type:         "comfyui",
		TaskType:     task.TaskType,
		Name:         task.Name,
		Description:  task.Description,
		ResultURL:    resultURL,
		InputParams:  inputParams,
		OutputParams: outputParams,
		UserID:       task.UserID,
		Username:     username,
		CreatedAt:    task.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// 从数字人任务转换为响应结构体
func convertDigitalHumanToResponse(dh models.DigitalHuman) InspirationTaskResponse {
	// 获取用户名
	username := "用户"
	if dh.UserID > 0 {
		var user models.User
		db.DB.Select("username").Where("id = ?", dh.UserID).First(&user)
		if user.Username != "" {
			username = user.Username
		}
	}

	// 适配数字人任务为与图像处理任务相同的格式
	inputParams := []InspirationParam{
		{Key: "audio_file", Label: "音频文件", Type: "audio", Value: utils.GetFileURL(dh.AudioURL)},
		{Key: "video_file", Label: "视频文件", Type: "video", Value: utils.GetFileURL(dh.VideoURL)},
	}

	// 添加其他参数
	if dh.Chaofen > 0 {
		inputParams = append(inputParams, InspirationParam{
			Key: "chaofen", Label: "超分", Type: "text", Value: "开启",
		})
	}

	if dh.WatermarkSwitch > 0 {
		inputParams = append(inputParams, InspirationParam{
			Key: "watermark", Label: "水印", Type: "text", Value: "开启",
		})
	}

	inputParams = append(inputParams, InspirationParam{
		Key: "pn", Label: "PN值", Type: "text", Value: utils.ToJSONString(dh.PN),
	})

	// 设置输出参数
	outputParams := []InspirationParam{
		{Key: "result_video", Label: "合成视频", Type: "video", Value: utils.GetFileURL(dh.ResultURL)},
	}

	return InspirationTaskResponse{
		ID:           dh.ID,
		Type:         "digital_human",
		TaskType:     "digital_human",
		Name:         dh.Name,
		Description:  dh.Description,
		ResultURL:    utils.GetFileURL(dh.ResultURL),
		InputParams:  inputParams,
		OutputParams: outputParams,
		UserID:       dh.UserID,
		Username:     username,
		CreatedAt:    dh.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
