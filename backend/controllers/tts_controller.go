package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"VirtualHumanStudio/backend/config"
	"VirtualHumanStudio/backend/db"
	"VirtualHumanStudio/backend/models"
	"VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TTSRequest TTS请求
type TTSRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	Type        string `form:"type" binding:"required,oneof=text2speech speech2text"` // 任务类型
	InputText   string `form:"input_text"`                                            // 文本转语音时的输入文本
	SpeakerName string `form:"speaker_name"`                                          // 使用的音色名称
	// 语音转文本时，文件通过multipart/form-data上传
}

// TTSResponse TTS响应
type TTSResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	InputText   string    `json:"input_text,omitempty"`
	InputFile   string    `json:"input_file,omitempty"`
	OutputFile  string    `json:"output_file,omitempty"`
	SpeakerName string    `json:"speaker_name,omitempty"`
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
		Type:        req.Type,
		Status:      "pending",
	}

	// 根据任务类型处理不同的输入
	switch req.Type {
	case "text2speech":
		// 文本转语音
		if req.InputText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文本转语音任务需要提供输入文本"})
			return
		}
		if req.SpeakerName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文本转语音任务需要提供音色名称"})
			return
		}

		ttsTask.InputText = req.InputText
		ttsTask.SpeakerName = req.SpeakerName

	case "speech2text":
		// 语音转文本
		file, err := c.FormFile("input_file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "语音转文本任务需要提供音频文件"})
			return
		}

		// 检查文件类型
		ext := filepath.Ext(file.Filename)
		if ext != ".wav" && ext != ".mp3" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持WAV或MP3格式的音频文件"})
			return
		}

		// 生成唯一文件名
		uniqueID := uuid.New().String()
		fileName := fmt.Sprintf("%s%s", uniqueID, ext)
		filePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, fileName)
		fullFilePath := utils.GetFilePath(config.AppConfig.DataDir, fileName)

		// 保存文件
		if err := c.SaveUploadedFile(file, fullFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
			return
		}

		ttsTask.InputFile = filePath

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的任务类型"})
		return
	}

	// 保存任务记录
	result := db.DB.Create(&ttsTask)
	if result.Error != nil {
		// 如果有上传的文件，删除它
		if ttsTask.InputFile != "" {
			os.Remove(ttsTask.InputFile)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建TTS任务记录失败: " + result.Error.Error()})
		return
	}

	// 异步处理TTS任务
	go processTTSTask(ttsTask)

	// 返回响应
	c.JSON(http.StatusCreated, gin.H{
		"message": "TTS任务已创建",
		"tts_task": TTSResponse{
			ID:          ttsTask.ID,
			Name:        ttsTask.Name,
			Description: ttsTask.Description,
			Type:        ttsTask.Type,
			InputText:   ttsTask.InputText,
			InputFile:   utils.GetFileURL(ttsTask.InputFile),
			OutputFile:  utils.GetFileURL(ttsTask.OutputFile),
			SpeakerName: ttsTask.SpeakerName,
			Status:      ttsTask.Status,
			CreatedAt:   ttsTask.CreatedAt,
			UpdatedAt:   ttsTask.UpdatedAt,
		},
	})
}

// processTTSTask 处理TTS任务
func processTTSTask(task models.TTSTask) {
	// 更新状态为处理中
	db.DB.Model(&task).Update("status", "processing")

	// 构建API请求
	apiReq := APITTSRequest{
		ModelName:   "CosyVoice2-0.5B_1",
		SpeakerName: task.SpeakerName,
		Text:        task.InputText,
		Language:    "mandarin",
		SpkRate:     1.0,
	}

	// 序列化请求
	reqData, err := json.Marshal(apiReq)
	if err != nil {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "序列化请求失败: " + err.Error(),
		})
		return
	}

	// 发送请求
	resp, err := http.Post(config.AppConfig.TTSAPI, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "调用API失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "读取API响应失败: " + err.Error(),
		})
		return
	}

	// 解析响应
	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "解析API响应失败: " + err.Error(),
		})
		return
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": fmt.Sprintf("API返回错误: %d %s", resp.StatusCode, string(respBody)),
		})
		return
	}

	// 处理响应结果
	switch task.Type {
	case "text2speech":
		// 文本转语音，保存输出文件
		waveBase64, ok := apiResp["wave_base64"].(string)
		if !ok {
			db.DB.Model(&task).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "API响应中未包含音频数据",
			})
			return
		}

		// 生成唯一的输出文件名
		outputFileName := fmt.Sprintf("%s.wav", uuid.New().String())
		outputFilePath := utils.GetUserFilePath(task.UserID, config.AppConfig.AudioDir, outputFileName)
		fullOutputFilePath := utils.GetFilePath(config.AppConfig.DataDir, outputFilePath)
		// 将Base64音频数据解码并保存为文件
		_, err := utils.Base64ToFile(waveBase64, fullOutputFilePath)
		if err != nil {
			db.DB.Model(&task).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "保存音频文件失败: " + err.Error(),
			})
			return
		}

		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":      "completed",
			"output_file": outputFilePath,
		})

	case "speech2text":
		// 语音转文本，保存输出文本
		outputText, ok := apiResp["output_text"].(string)
		if !ok {
			db.DB.Model(&task).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "API响应中未包含输出文本",
			})
			return
		}

		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":     "completed",
			"input_text": outputText, // 将识别结果保存到input_text字段
		})
	}
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
			Type:        ttsTask.Type,
			InputText:   ttsTask.InputText,
			InputFile:   utils.GetFileURL(ttsTask.InputFile),
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
			Type:        task.Type,
			InputText:   task.InputText,
			InputFile:   utils.GetFileURL(task.InputFile),
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
	if ttsTask.InputFile != "" {
		fullInputFilePath := utils.GetFilePath(config.AppConfig.DataDir, ttsTask.InputFile)
		os.Remove(fullInputFilePath)
	}
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
