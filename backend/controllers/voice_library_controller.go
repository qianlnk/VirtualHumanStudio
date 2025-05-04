package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// UploadVoice 上传音色文件到音色库
func UploadVoice(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 解析表单数据
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "音色名称不能为空"})
		return
	}

	description := c.PostForm("description")
	modelName := c.PostForm("model_name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模型名称不能为空"})
		return
	}

	isPublicStr := c.DefaultPostForm("is_public", "false")
	isPublic := isPublicStr == "true"

	// 获取音色模型文件
	modelFile, err := c.FormFile("model_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供音色模型文件"})
		return
	}

	// 获取试听音频文件（可选）
	sampleFile, _ := c.FormFile("sample_file")

	// 检查音色名称是否已存在
	var existingVoice models.VoiceLibrary
	result := db.DB.Where("name = ?", name).First(&existingVoice)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "音色名称已存在"})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 生成唯一文件名并保存音色模型文件
	modelUniqueID := uuid.New().String()
	modelFileName := fmt.Sprintf("%s%s", modelUniqueID, filepath.Ext(modelFile.Filename))
	modelFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, modelFileName)
	modelFullPath := filepath.Join(config.AppConfig.DataDir, modelFilePath)

	if err := c.SaveUploadedFile(modelFile, modelFullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存音色模型文件失败: " + err.Error()})
		return
	}

	// 保存试听音频文件（如果有）
	var sampleFilePath string
	if sampleFile != nil {
		ext := filepath.Ext(sampleFile.Filename)
		if ext != ".wav" && ext != ".mp3" {
			// 删除已上传的模型文件
			os.Remove(modelFullPath)
			c.JSON(http.StatusBadRequest, gin.H{"error": "试听音频文件仅支持WAV或MP3格式"})
			return
		}

		sampleUniqueID := uuid.New().String()
		sampleFileName := fmt.Sprintf("%s%s", sampleUniqueID, ext)
		sampleFilePath = utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, sampleFileName)
		sampleFullPath := filepath.Join(config.AppConfig.DataDir, sampleFilePath)

		if err := c.SaveUploadedFile(sampleFile, sampleFullPath); err != nil {
			// 删除已上传的模型文件
			os.Remove(modelFullPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存试听音频文件失败: " + err.Error()})
			return
		}
	}

	// 创建音色库记录
	voiceLibrary := models.VoiceLibrary{
		Name:        name,
		Description: description,
		ModelName:   modelName,
		ModelFile:   modelFilePath,
		SampleFile:  sampleFilePath,
		Type:        "original",
		OwnerID:     userID.(uint),
		IsPublic:    isPublic,
	}

	result = db.DB.Create(&voiceLibrary)
	if result.Error != nil {
		// 删除已上传的文件
		os.Remove(modelFullPath)
		if sampleFilePath != "" {
			os.Remove(filepath.Join(config.AppConfig.DataDir, sampleFilePath))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建音色库记录失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "音色上传成功",
		"voice": gin.H{
			"id":          voiceLibrary.ID,
			"name":        voiceLibrary.Name,
			"description": voiceLibrary.Description,
			"model_name":  voiceLibrary.ModelName,
			"model_file":  utils.GetFileURL(voiceLibrary.ModelFile),
			"sample_file": utils.GetFileURL(voiceLibrary.SampleFile),
			"type":        voiceLibrary.Type,
			"owner_id":    voiceLibrary.OwnerID,
			"is_public":   voiceLibrary.IsPublic,
			"created_at":  voiceLibrary.CreatedAt,
		},
	})
}

// ListVoices 获取音色库列表
func ListVoices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 查询条件：用户自己的音色或公开的音色
	query := db.DB.Where("owner_id = ? OR is_public = ?", userID, true)

	// 查询总数
	var count int64
	query.Model(&models.VoiceLibrary{}).Count(&count)

	count += int64(len(config.AppConfig.OfficialVoices))

	// 查询列表
	var voices []models.VoiceLibrary
	result := query.Order("created_at DESC").Find(&voices)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 构建响应
	responses := make([]map[string]interface{}, len(voices))
	for i, v := range voices {
		responses[i] = map[string]interface{}{
			"id":          v.ID,
			"name":        v.Name,
			"alias":       v.Name, // 别名字段，用于前端显示，例如 "官方音色"
			"description": v.Description,
			"model_name":  v.ModelName,
			"model_file":  utils.GetFileURL(v.ModelFile),
			"sample_file": utils.GetFileURL(v.SampleFile),
			"type":        v.Type,
			"owner_id":    v.OwnerID,
			"is_public":   v.IsPublic,
			"created_at":  v.CreatedAt,
			"is_owner":    v.OwnerID == userID.(uint),
		}
	}

	// 添加官方音色
	layout := "2006-01-02 15:04:05"
	createdAt, _ := time.Parse(layout, "2025-01-02 15:04:05")
	for _, officialVoice := range config.AppConfig.OfficialVoices {
		responses = append(responses, map[string]interface{}{
			"id":          officialVoice.ID,
			"name":        officialVoice.ID,
			"alias":       officialVoice.Alias, // 别名字段，用于前端显示，例如 "官方音色
			"description": officialVoice.Description,
			"type":        "official",
			"owner_id":    0,
			"is_public":   true,
			"is_owner":    false,
			"gender":      officialVoice.Gender,
			"created_at":  createdAt,
			"sample_file": utils.GetFileURL(officialVoice.SampleFile),
		})
	}

	start := (page - 1) * size
	end := start + size
	if start >= len(responses) {
		responses = []map[string]interface{}{}
	} else if end > len(responses) {
		end = len(responses)
	}

	responses = responses[start:end]

	c.JSON(http.StatusOK, gin.H{
		"total":  count,
		"page":   page,
		"size":   size,
		"voices": responses,
	})
}

// DeleteVoice 删除音色
func DeleteVoice(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取音色ID
	id := c.Param("id")

	// 查询音色
	var voice models.VoiceLibrary
	result := db.DB.First(&voice, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "音色不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if voice.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此音色"})
		return
	}

	// 删除音色模型文件和试听音频文件
	if voice.ModelFile != "" {
		modelFullPath := filepath.Join(config.AppConfig.DataDir, voice.ModelFile)
		os.Remove(modelFullPath)
	}
	if voice.SampleFile != "" {
		sampleFullPath := filepath.Join(config.AppConfig.DataDir, voice.SampleFile)
		os.Remove(sampleFullPath)
	}

	// 删除记录
	result = db.DB.Delete(&voice)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "音色已删除"})
}
