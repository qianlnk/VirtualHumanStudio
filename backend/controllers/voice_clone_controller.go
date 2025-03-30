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

// VoiceCloneRequest 音色克隆请求
type VoiceCloneRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	ModelName   string `form:"model_name" binding:"required"`
	PromptText  string `form:"prompt_text" binding:"required"`
	SpeakerName string `form:"speaker_name" binding:"required"`
	// 文件通过multipart/form-data上传
}

// VoiceCloneResponse 音色克隆响应
type VoiceCloneResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ModelName   string    `json:"model_name"`
	PromptFile  string    `json:"prompt_file"`
	PromptText  string    `json:"prompt_text"`
	SpeakerName string    `json:"speaker_name"`
	Status      string    `json:"status"`
	Result      string    `json:"result"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APIVoiceCloneRequest API音色克隆请求
type APIVoiceCloneRequest struct {
	ModelName   string `json:"model_name"`
	PromptFile  string `json:"prompt_file"`
	PromptText  string `json:"prompt_text"`
	SpeakerName string `json:"speaker_name"`
}

// CreateVoiceClone 创建音色克隆任务
func CreateVoiceClone(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 解析表单数据
	var req VoiceCloneRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("prompt_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供音频文件"})
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
	filePath := utils.GetFilePath(config.AppConfig.UploadDir, fileName)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	// 创建音色克隆记录
	voiceClone := models.VoiceClone{
		UserID:      userID.(uint),
		Name:        req.Name,
		Description: req.Description,
		ModelName:   req.ModelName,
		PromptFile:  filePath,
		PromptText:  req.PromptText,
		SpeakerName: req.SpeakerName,
		Status:      "pending",
	}

	result := db.DB.Create(&voiceClone)
	if result.Error != nil {
		// 删除已上传的文件
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建音色克隆记录失败: " + result.Error.Error()})
		return
	}

	// 调用音色克隆API
	go func() {
		// 更新状态为处理中
		db.DB.Model(&voiceClone).Update("status", "processing")

		// 构建API请求
		apiReq := APIVoiceCloneRequest{
			ModelName:   req.ModelName,
			PromptFile:  filePath, // 这里应该是服务器上的路径，可能需要调整
			PromptText:  req.PromptText,
			SpeakerName: req.SpeakerName,
		}

		// 重用现有的API调用逻辑
		handleVoiceCloneAPI(apiReq, &voiceClone)
	}()

	// 返回响应
	c.JSON(http.StatusCreated, gin.H{
		"message": "音色克隆任务已创建",
		"voice_clone": VoiceCloneResponse{
			ID:          voiceClone.ID,
			Name:        voiceClone.Name,
			Description: voiceClone.Description,
			ModelName:   voiceClone.ModelName,
			PromptFile:  voiceClone.PromptFile,
			PromptText:  voiceClone.PromptText,
			SpeakerName: voiceClone.SpeakerName,
			Status:      voiceClone.Status,
			Result:      voiceClone.Result,
			CreatedAt:   voiceClone.CreatedAt,
			UpdatedAt:   voiceClone.UpdatedAt,
		},
	})
}

// GetVoiceClone 获取音色克隆任务
func GetVoiceClone(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var voiceClone models.VoiceClone
	result := db.DB.First(&voiceClone, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "音色克隆任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if voiceClone.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此任务"})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"voice_clone": VoiceCloneResponse{
			ID:          voiceClone.ID,
			Name:        voiceClone.Name,
			Description: voiceClone.Description,
			ModelName:   voiceClone.ModelName,
			PromptFile:  voiceClone.PromptFile,
			PromptText:  voiceClone.PromptText,
			SpeakerName: voiceClone.SpeakerName,
			Status:      voiceClone.Status,
			Result:      voiceClone.Result,
			CreatedAt:   voiceClone.CreatedAt,
			UpdatedAt:   voiceClone.UpdatedAt,
		},
	})
}

// ListVoiceClones 获取音色克隆任务列表
func ListVoiceClones(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 查询任务总数
	var count int64
	db.DB.Model(&models.VoiceClone{}).Where("user_id = ?", userID).Count(&count)

	// 查询任务列表
	var voiceClones []models.VoiceClone
	result := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&voiceClones)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 构建响应
	responses := make([]VoiceCloneResponse, len(voiceClones))
	for i, vc := range voiceClones {
		responses[i] = VoiceCloneResponse{
			ID:          vc.ID,
			Name:        vc.Name,
			Description: vc.Description,
			ModelName:   vc.ModelName,
			PromptFile:  vc.PromptFile,
			PromptText:  vc.PromptText,
			SpeakerName: vc.SpeakerName,
			Status:      vc.Status,
			Result:      vc.Result,
			CreatedAt:   vc.CreatedAt,
			UpdatedAt:   vc.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":        count,
		"page":         page,
		"size":         size,
		"voice_clones": responses,
	})
}

// DeleteVoiceClone 删除音色克隆任务
func DeleteVoiceClone(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var voiceClone models.VoiceClone
	result := db.DB.First(&voiceClone, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "音色克隆任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if voiceClone.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此任务"})
		return
	}

	// 删除关联文件
	if voiceClone.PromptFile != "" {
		os.Remove(voiceClone.PromptFile)
	}
	if voiceClone.Result != "" {
		os.Remove(voiceClone.Result)
	}

	// 删除记录
	result = db.DB.Delete(&voiceClone)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "音色克隆任务已删除"})
}

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
	isPublicStr := c.DefaultPostForm("is_public", "false")
	isPublic := isPublicStr == "true"

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供音频文件"})
		return
	}

	// 检查文件类型
	ext := filepath.Ext(file.Filename)
	if ext != ".wav" && ext != ".mp3" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持WAV或MP3格式的音频文件"})
		return
	}

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

	// 生成唯一文件名
	uniqueID := uuid.New().String()
	fileName := fmt.Sprintf("%s%s", uniqueID, ext)
	filePath := utils.GetFilePath(config.AppConfig.UploadDir, fileName)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	// 创建音色库记录
	voiceLibrary := models.VoiceLibrary{
		Name:        name,
		Description: description,
		FilePath:    filePath,
		Type:        "original",
		OwnerID:     userID.(uint),
		IsPublic:    isPublic,
	}

	result = db.DB.Create(&voiceLibrary)
	if result.Error != nil {
		// 删除已上传的文件
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建音色库记录失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "音色上传成功",
		"voice": gin.H{
			"id":          voiceLibrary.ID,
			"name":        voiceLibrary.Name,
			"description": voiceLibrary.Description,
			"file_path":   voiceLibrary.FilePath,
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

	// 查询列表
	var voices []models.VoiceLibrary
	result := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&voices)
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
			"description": v.Description,
			"file_path":   v.FilePath,
			"type":        v.Type,
			"owner_id":    v.OwnerID,
			"is_public":   v.IsPublic,
			"created_at":  v.CreatedAt,
			"is_owner":    v.OwnerID == userID.(uint),
		}
	}

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

	// 删除文件
	if voice.FilePath != "" {
		os.Remove(voice.FilePath)
	}

	// 删除记录
	result = db.DB.Delete(&voice)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "音色已删除"})
}

// DownloadVoice 下载音色文件
func DownloadVoice(c *gin.Context) {
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
	if !voice.IsPublic && voice.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权下载此音色"})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(voice.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "音色文件不存在"})
		return
	}

	// 获取文件名
	fileName := filepath.Base(voice.FilePath)

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.File(voice.FilePath)
}

// handleVoiceCloneAPI 处理音色克隆API调用
func handleVoiceCloneAPI(apiReq APIVoiceCloneRequest, voiceClone *models.VoiceClone) {
	// 序列化请求
	reqData, err := json.Marshal(apiReq)
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "序列化请求失败: " + err.Error(),
		})
		return
	}

	// 发送请求
	resp, err := http.Post(config.AppConfig.VoiceCloneAPI, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "调用API失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "读取API响应失败: " + err.Error(),
		})
		return
	}

	// 解析响应
	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "解析API响应失败: " + err.Error(),
		})
		return
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": fmt.Sprintf("API返回错误: %d %s", resp.StatusCode, string(respBody)),
		})
		return
	}

	// 假设API返回了结果文件路径
	resultFile, ok := apiResp["result_file"].(string)
	if !ok {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":  "completed",
			"task_id": apiResp["task_id"],
			"result":  "", // 可能需要后续查询结果
		})
	} else {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":  "completed",
			"task_id": apiResp["task_id"],
			"result":  resultFile,
		})
	}
}
