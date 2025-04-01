package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"VirtualHumanStudio/backend/config"
	"VirtualHumanStudio/backend/db"
	"VirtualHumanStudio/backend/models"

	"github.com/gin-gonic/gin"
)

// 支持的媒体类型
var mediaTypes = map[string]string{
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
}

func FileView(c *gin.Context) {
	// 获取文件路径参数
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供文件路径"})
		return
	}
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 构建完整文件路径
	fullPath := filepath.Join(config.AppConfig.DataDir, filePath)
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 检查文件访问权限
	if !checkFileAccess(uint(userID.(uint)), filePath) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该文件"})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
	// 发送文件
	c.File(fullPath)
}

// checkFileAccess 检查用户是否有权限访问文件
func checkFileAccess(userID uint, filePath string) bool {
	// 检查文件是否属于用户的音色克隆
	var voiceClone models.VoiceClone
	result := db.DB.Where("user_id = ? AND (prompt_file = ? OR result = ?)", userID, filePath, filePath).First(&voiceClone)
	if result.Error == nil {
		return true
	}

	// 检查文件是否属于用户的TTS任务
	var ttsTask models.TTSTask
	result = db.DB.Where("user_id = ? AND (output_file = ? OR input_file = ?)", userID, filePath, filePath).First(&ttsTask)
	if result.Error == nil {
		return true
	}

	// 检查文件是否属于用户的数字人
	var digitalHuman models.DigitalHuman
	result = db.DB.Where("user_id = ? AND (audio_url = ? OR video_url = ? OR result_url = ?)", userID, filePath, filePath, filePath).First(&digitalHuman)
	if result.Error == nil {
		return true
	}

	// 检查文件是否属于用户的音色库
	var voiceLibrary models.VoiceLibrary
	result = db.DB.Where("owner_id = ? AND (model_file = ? or sample_file = ?)", userID, filePath, filePath).First(&voiceLibrary)
	if result.Error == nil {
		return true
	}

	return false
}
