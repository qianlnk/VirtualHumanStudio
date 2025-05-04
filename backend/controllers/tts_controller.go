package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// TTSRequest TTS请求
type TTSRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	InputText   string `form:"input_text" binding:"required"`   // 输入文本
	SpeakerName string `form:"speaker_name" binding:"required"` // 使用的音色名称
}

// TTSResponse TTS响应
type TTSResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	InputText   string    `json:"input_text"`
	OutputFile  string    `json:"output_file,omitempty"`
	SpeakerName string    `json:"speaker_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APITTSRequest API TTS请求
type APITTSRequest struct {
	ModelName   string  `json:"model_name"`   // 模型名称
	SpeakerName string  `json:"speaker_name"` // 使用的音色名称
	Text        string  `json:"text"`         // 输入文本
	Language    string  `json:"language"`     // 语言，如mandarin
	SpkRate     float64 `json:"spk_rate"`     // 语速
}

// CreateTTSTask 创建TTS任务
func CreateTTSTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 解析表单数据
	var req TTSRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 创建TTS任务记录
	ttsTask := models.TTSTask{
		UserID:      userID.(uint),
		Name:        req.Name,
		Description: req.Description,
		InputText:   req.InputText,
		SpeakerName: req.SpeakerName,
		Status:      "pending",
	}

	// 保存到数据库
	if err := db.DB.Create(&ttsTask).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败: " + err.Error()})
		return
	}

	// 异步处理TTS任务
	ctx := context.WithValue(context.Background(), "user_id", userID)
	go processTTSTask(ctx, ttsTask)

	// 返回响应
	c.JSON(http.StatusCreated, gin.H{
		"message": "TTS任务已创建",
		"tts_task": TTSResponse{
			ID:          ttsTask.ID,
			Name:        ttsTask.Name,
			Description: ttsTask.Description,
			InputText:   ttsTask.InputText,
			OutputFile:  utils.GetFileURL(ttsTask.OutputFile),
			SpeakerName: ttsTask.SpeakerName,
			Status:      ttsTask.Status,
			CreatedAt:   ttsTask.CreatedAt,
			UpdatedAt:   ttsTask.UpdatedAt,
		},
	})
}

// processTTSTask 处理TTS任务
func processTTSTask(ctx context.Context, task models.TTSTask) {
	// 更新状态为处理中
	db.DB.Model(&task).Update("status", "processing")

	modelName := "CosyVoice2-0.5B_1"
	language := "mandarin"
	speakerName := task.SpeakerName
	official := false
	if v, ok := config.AppConfig.OfficialVoices[task.SpeakerName]; ok {
		modelName = v.Model
		language = "auto"
		official = true
		speakerName = v.TimbreID
	} else {
		speakerName = fmt.Sprintf("%d_%s", task.UserID, task.SpeakerName)
	}

	// 构建API请求
	apiReq := APITTSRequest{
		ModelName:   modelName,
		SpeakerName: speakerName,
		Text:        task.InputText,
		Language:    language,
		SpkRate:     1.0,
	}

	outputFilePath, err := ttsInvoke(ctx, &apiReq, official)
	if err != nil {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": err.Error(),
		})
		return
	}

	db.DB.Model(&task).Updates(map[string]interface{}{
		"status":      "completed",
		"output_file": outputFilePath,
	})
}

// GetTTSTask 获取TTS任务
func GetTTSTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var ttsTask models.TTSTask
	result := db.DB.First(&ttsTask, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "TTS任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if ttsTask.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此任务"})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"tts_task": TTSResponse{
			ID:          ttsTask.ID,
			Name:        ttsTask.Name,
			Description: ttsTask.Description,
			InputText:   ttsTask.InputText,
			OutputFile:  utils.GetFileURL(ttsTask.OutputFile),
			SpeakerName: ttsTask.SpeakerName,
			Status:      ttsTask.Status,
			CreatedAt:   ttsTask.CreatedAt,
			UpdatedAt:   ttsTask.UpdatedAt,
		},
	})
}

// ListTTSTasks 获取TTS任务列表
func ListTTSTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 查询任务总数
	var count int64
	db.DB.Model(&models.TTSTask{}).Where("user_id = ?", userID).Count(&count)

	// 查询任务列表
	var ttsTasks []models.TTSTask
	result := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&ttsTasks)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 构建响应
	responses := make([]TTSResponse, len(ttsTasks))
	for i, task := range ttsTasks {
		responses[i] = TTSResponse{
			ID:          task.ID,
			Name:        task.Name,
			Description: task.Description,
			InputText:   task.InputText,
			OutputFile:  utils.GetFileURL(task.OutputFile),
			SpeakerName: task.SpeakerName,
			Status:      task.Status,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     count,
		"page":      page,
		"size":      size,
		"tts_tasks": responses,
	})
}

// DeleteTTSTask 删除TTS任务
func DeleteTTSTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var ttsTask models.TTSTask
	result := db.DB.First(&ttsTask, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "TTS任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if ttsTask.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此任务"})
		return
	}

	// 删除关联文件
	if ttsTask.OutputFile != "" {
		fullOutputFilePath := utils.GetFilePath(config.AppConfig.DataDir, ttsTask.OutputFile)
		os.Remove(fullOutputFilePath)
	}

	// 删除记录
	result = db.DB.Delete(&ttsTask)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "TTS任务已删除"})
}
