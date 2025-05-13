package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
)

// ShareTaskResponseData 共享任务响应数据
type ShareTaskResponseData struct {
	ID           uint                 `json:"id"`                                            // 分享ID
	Mode         string               `json:"mode"`                                          // 任务类型: "comfyui" 或 "digital_human"
	TaskType     string               `json:"task_type"`                                     // 具体任务类型，如"accessory"、"workflow"等
	TaskID       uint                 `json:"task_id"`                                       // 任务ID
	Name         string               `json:"name"`                                          // 任务名称
	Description  string               `json:"description"`                                   // 任务描述
	ResultURL    string               `json:"result_url"`                                    // 结果图片/视频URL
	InputParams  []*models.InputParam `json:"input_params"`                                  // 输入参数
	OutputParams []*models.InputParam `json:"output_params"`                                 // 输出参数
	UserID       uint                 `json:"user_id"`                                       // 创建者ID
	Username     string               `json:"username"`                                      // 创建者用户名
	CreatedAt    string               `json:"created_at"`                                    // 创建时间
	ShareStatus  string               `json:"share_status" gorm:"size:20;default:'private'"` // private, pending_review, approved, rejected
}

// ShareTaskRequest 分享任务请求
type ShareTaskRequest struct {
	TaskID   uint   `json:"task_id" binding:"required"`
	Mode     string `json:"mode" binding:"required"`      // comfyui 或 digital_human
	TaskType string `json:"task_type" binding:"required"` // 任务类型
}

// ShareTaskResponse 分享任务响应
type ShareTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ReviewTaskRequest 审核任务请求
type ReviewTaskRequest struct {
	ShareID      uint   `json:"share_id" binding:"required"`  // 分享ID
	TaskID       uint   `json:"task_id" binding:"required"`   // 任务ID
	Mode         string `json:"mode" binding:"required"`      // comfyui 或 digital_human
	TaskType     string `json:"task_type" binding:"required"` // 任务类型
	Status       string `json:"status" binding:"required"`    // approved 或 rejected
	RejectReason string `json:"reject_reason"`                // 拒绝原因
}

// ReviewTaskResponse 审核任务响应
type ReviewTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ShareTask 分享任务
func ShareTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req ShareTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	now := time.Now()

	shareTask := models.ShareTask{
		Mode:        req.Mode,
		TaskType:    req.TaskType,
		ShareStatus: "pending_review",
		ShareTime:   &now,
	}

	if req.Mode == "digital_human" {
		// 查询数字人任务
		var task models.DigitalHuman
		if err := db.DB.Where("id = ? AND user_id = ?", req.TaskID, userID).First(&task).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}

		// 检查任务状态
		if task.Status != "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "只能分享已完成的任务"})
			return
		}

		shareTask.TaskID = task.ID

		if err := db.DB.Create(&shareTask).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分享任务失败: " + err.Error()})
			return
		}

		if err := db.DB.Model(&task).Updates(map[string]interface{}{
			"share_status": "pending_review",
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, ShareTaskResponse{
			Success: true,
			Message: "任务已提交审核",
		})
	} else {
		// 查询ComfyUI任务
		var task models.ComfyUIWorkflowTask
		if err := db.DB.Where("id = ? AND user_id = ?", req.TaskID, userID).First(&task).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}

		// 检查任务状态
		if task.Status != "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "只能分享已完成的任务"})
			return
		}

		shareTask.TaskID = task.ID

		if err := db.DB.Create(&shareTask).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分享任务失败: " + err.Error()})
			return
		}

		// 更新分享状态
		if err := db.DB.Model(&task).Updates(map[string]interface{}{
			"share_status": "pending_review",
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, ShareTaskResponse{
			Success: true,
			Message: "任务已提交审核",
		})
	}
}

// ReviewTask 审核任务
func ReviewTask(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req ReviewTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	now := time.Now()

	if req.Mode == "digital_human" {
		// 查询数字人任务
		var task models.DigitalHuman
		if err := db.DB.Where("id = ?", req.TaskID).First(&task).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}

		// // 检查任务状态
		// if task.ShareStatus != "pending_review" {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "任务不处于待审核状态"})
		// 	return
		// }

		// 更新分享状态
		if err := db.DB.Model(&task).Update("share_status", req.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + err.Error()})
			return
		}

		// 查询分享任务
		var shareTask models.ShareTask
		if err := db.DB.Where("id = ?", req.ShareID).First(&shareTask).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "分享任务不存在"})
			return
		}

		// 构建更新字段
		updateFields := map[string]interface{}{
			"share_status": req.Status,
			"reviewer_id":  adminID,
			"review_time":  &now,
		}

		// 如果是拒绝，添加拒绝原因
		if req.Status == "rejected" {
			updateFields["reject_reason"] = req.RejectReason
		}

		// 打印调试信息
		fmt.Printf("更新分享任务ID: %d, 字段: %+v\n", shareTask.ID, updateFields)

		// 直接使用已查询到的shareTask对象进行更新
		if err := db.DB.Model(&shareTask).Updates(updateFields).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, ReviewTaskResponse{
			Success: true,
			Message: "任务审核完成",
		})
	} else {
		// 查询ComfyUI任务
		var task models.ComfyUIWorkflowTask
		if err := db.DB.Where("id = ?", req.TaskID).First(&task).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}

		// // 检查任务状态
		// if task.ShareStatus != "pending_review" {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "任务不处于待审核状态"})
		// 	return
		// }

		// 更新分享状态
		if err := db.DB.Model(&task).Update("share_status", req.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + err.Error()})
			return
		}

		// 查询分享任务
		var shareTask models.ShareTask
		if err := db.DB.Where("id = ?", req.ShareID).First(&shareTask).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "分享任务不存在"})
			return
		}

		// 构建更新字段
		updateFields := map[string]interface{}{
			"share_status": req.Status,
			"reviewer_id":  adminID,
			"review_time":  &now,
		}

		// 如果是拒绝，添加拒绝原因
		if req.Status == "rejected" {
			updateFields["reject_reason"] = req.RejectReason
		}

		// 打印调试信息
		fmt.Printf("更新分享任务ID: %d, 字段: %+v\n", shareTask.ID, updateFields)

		// 直接使用已查询到的shareTask对象进行更新
		if err := db.DB.Model(&shareTask).Updates(updateFields).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, ReviewTaskResponse{
			Success: true,
			Message: "任务审核完成",
		})
	}
}

// GetPendingReviewTasks 获取待审核任务
func GetPendingReviewTasks(c *gin.Context) {
	// 分页参数
	page, size := utils.GetPaginationParams(c)
	shareStatus := c.Query("share_status")

	var shareTasks []models.ShareTask
	var count int64
	var responses []ShareTaskResponseData

	query := db.DB.Model(&models.ShareTask{}).Where("share_status =?", "pending_review")
	if shareStatus != "" {
		query = query.Where("share_status = ?", shareStatus)
	}

	query.Count(&count)
	query.Order("id DESC").
		Offset((page - 1) * size).
		Limit(1).
		Find(&shareTasks)

	modIds := map[string][]uint{}
	shareIDs := map[string]uint{}
	for _, task := range shareTasks {
		modIds[task.Mode] = append(modIds[task.Mode], task.TaskID)
		shareIDs[task.Mode+"_"+strconv.Itoa(int(task.TaskID))] = task.ID
	}

	for mode, ids := range modIds {
		switch mode {
		case "comfyui":
			var tasks []models.ComfyUIWorkflowTask
			db.DB.Where("id IN (?)", ids).Find(&tasks)
			for _, task := range tasks {
				var inputParams []*models.InputParam
				var outputParams []*models.InputParam

				json.Unmarshal([]byte(task.InputParams), &inputParams)
				json.Unmarshal([]byte(task.OutputParams), &outputParams)

				for _, param := range inputParams {
					if param.Type == "image" || param.Type == "mask" || param.Type == "video" || param.Type == "audio" {
						param.Value = utils.GetFileURL(param.Value)
					}
				}

				for _, param := range outputParams {
					if param.Type == "image" || param.Type == "mask" || param.Type == "video" || param.Type == "audio" {
						param.Value = utils.GetFileURL(param.Value)
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

				responses = append(responses, ShareTaskResponseData{
					ID:           shareIDs[mode+"_"+strconv.Itoa(int(task.ID))],
					TaskID:       task.ID,
					Mode:         "comfyui",
					TaskType:     task.TaskType,
					Name:         task.Name,
					Description:  task.Description,
					InputParams:  inputParams,
					OutputParams: outputParams,
					UserID:       task.UserID,
					Username:     username,
					CreatedAt:    task.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
		case "digital_human":
			var tasks []models.DigitalHuman
			db.DB.Where("id IN (?)", ids).Find(&tasks)
			for _, task := range tasks {

				// 获取用户名
				username := "用户"
				if task.UserID > 0 {
					var user models.User
					db.DB.Select("username").Where("id = ?", task.UserID).First(&user)
					if user.Username != "" {
						username = user.Username
					}
				}

				// 适配数字人任务为与图像处理任务相同的格式
				inputParams := []*models.InputParam{
					{Key: "audio_file", Alias: "音频文件", Type: "audio", Value: utils.GetFileURL(task.AudioURL)},
					{Key: "video_file", Alias: "视频文件", Type: "video", Value: utils.GetFileURL(task.VideoURL)},
				}

				// 添加其他参数
				if task.Chaofen > 0 {
					inputParams = append(inputParams, &models.InputParam{
						Key: "chaofen", Alias: "超分", Type: "text", Value: "开启",
					})
				} else {
					inputParams = append(inputParams, &models.InputParam{
						Key: "chaofen", Alias: "超分", Type: "text", Value: "关闭",
					})
				}

				if task.WatermarkSwitch > 0 {
					inputParams = append(inputParams, &models.InputParam{
						Key: "watermark", Alias: "水印", Type: "text", Value: "开启",
					})
				} else {
					inputParams = append(inputParams, &models.InputParam{
						Key: "watermark", Alias: "水印", Type: "text", Value: "关闭",
					})
				}

				inputParams = append(inputParams, &models.InputParam{
					Key: "pn", Alias: "PN值", Type: "text", Value: fmt.Sprintf("%v", task.PN),
				})

				// 如果有任务代码，添加到输入参数中
				if task.TaskCode != "" {
					inputParams = append(inputParams, &models.InputParam{
						Key: "task_code", Alias: "任务代码", Type: "text", Value: task.TaskCode,
					})
				}

				// 设置输出参数
				outputParams := []*models.InputParam{
					{Key: "result_video", Alias: "合成视频", Type: "video", Value: utils.GetFileURL(task.ResultURL)},
				}

				// 如果有详细的任务记录等信息，添加到输出参数
				if task.ErrorMsg != "" {
					outputParams = append(outputParams, &models.InputParam{
						Key: "error_msg", Alias: "错误信息", Type: "text", Value: task.ErrorMsg,
					})
				}

				// 添加任务状态
				outputParams = append(outputParams, &models.InputParam{
					Key: "status", Alias: "任务状态", Type: "text", Value: getStatusText(task.Status),
				})

				responses = append(responses, ShareTaskResponseData{
					ID:           uint(shareIDs[mode+"_"+strconv.Itoa(int(task.ID))]),
					TaskID:       task.ID,
					Mode:         "digital_human",
					TaskType:     "digital_human",
					Name:         task.Name,
					Description:  task.Description,
					InputParams:  inputParams,
					OutputParams: outputParams,
					UserID:       task.UserID,
					Username:     username,
					CreatedAt:    task.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}

		}

		c.JSON(http.StatusOK, gin.H{
			"tasks": responses,
			"total": count,
			"page":  page,
			"size":  size,
		})
	}
}

func getStatusText(status string) string {
	switch status {
	case "completed":
		return "已完成"
	case "processing":
		return "处理中"
	case "pending_review":
		return "待审核"
	case "rejected":
		return "已拒绝"
	default:
		return "未知状态"
	}
}
