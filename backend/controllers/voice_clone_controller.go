package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
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
	// 生成相对路径和完整路径
	filePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, fileName)
	fullFilePath := filepath.Join(config.AppConfig.DataDir, filePath)

	// 保存文件
	if err := c.SaveUploadedFile(file, fullFilePath); err != nil {
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
		os.Remove(fullFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建音色克隆记录失败: " + result.Error.Error()})
		return
	}

	// 调用音色克隆API
	go func() {
		// 更新状态为处理中
		db.DB.Model(&voiceClone).Update("status", "processing")

		// 重用现有的API调用逻辑
		handleVoiceCloneAPI(&voiceClone)
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
			PromptFile:  utils.GetFileURL(voiceClone.PromptFile),
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
			PromptFile:  utils.GetFileURL(vc.PromptFile),
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
		fullPromptPath := filepath.Join(config.AppConfig.DataDir, voiceClone.PromptFile)
		os.Remove(fullPromptPath)
	}
	if voiceClone.Result != "" {
		fullResultPath := filepath.Join(config.AppConfig.DataDir, voiceClone.Result)
		os.Remove(fullResultPath)
	}

	// 删除记录
	result = db.DB.Delete(&voiceClone)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "音色克隆任务已删除"})
}

// AddVoiceToLibrary 添加音色到库中
func AddVoiceToLibrary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取音色克隆任务ID
	id := c.Param("id")

	// 查询音色克隆任务
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

	// 检查任务状态
	if voiceClone.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "音色克隆任务尚未完成"})
		return
	}

	// 检查音色文件是否存在
	if voiceClone.Result == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "音色文件不存在"})
		return
	}

	// 创建音色库记录
	voiceLibrary := models.VoiceLibrary{
		Name:        voiceClone.SpeakerName,
		Description: voiceClone.Description,
		ModelName:   voiceClone.ModelName,
		ModelFile:   voiceClone.Result,
		Type:        "cloned",
		OwnerID:     voiceClone.UserID,
		IsPublic:    false,
	}

	result = db.DB.Create(&voiceLibrary)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建音色库记录失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "音色已添加到库中",
		"voice": gin.H{
			"id":          voiceLibrary.ID,
			"name":        voiceLibrary.Name,
			"description": voiceLibrary.Description,
			"model_name":  voiceLibrary.ModelName,
			"model_file":  voiceLibrary.ModelFile,
			"type":        voiceLibrary.Type,
			"owner_id":    voiceLibrary.OwnerID,
			"is_public":   voiceLibrary.IsPublic,
			"created_at":  voiceLibrary.CreatedAt,
		},
	})
}

// handleVoiceCloneAPI 处理音色克隆API调用
func handleVoiceCloneAPI(voiceClone *models.VoiceClone) {
	// 1. 上传音频文件到音色克隆服务器
	fullPromptPath := filepath.Join(config.AppConfig.DataDir, voiceClone.PromptFile)
	file, err := os.Open(fullPromptPath)
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "打开音频文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	part, err := writer.CreateFormFile("attachment", filepath.Base(fullPromptPath))
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "创建表单失败: " + err.Error(),
		})
		return
	}
	if _, err = io.Copy(part, file); err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "复制文件失败: " + err.Error(),
		})
		return
	}

	// 添加path字段
	uploadPath := fmt.Sprintf("vhs/%d/audios", voiceClone.UserID)
	if err = writer.WriteField("path", uploadPath); err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "添加path字段失败: " + err.Error(),
		})
		return
	}
	writer.Close()

	// 发送文件上传请求
	uploadResp, err := http.Post(config.AppConfig.FileUploadAPI, writer.FormDataContentType(), body)
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "上传文件失败: " + err.Error(),
		})
		return
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "文件上传失败，状态码: " + fmt.Sprint(uploadResp.StatusCode),
		})
		return
	}

	// 2. 调用音色克隆API
	// 更新文件路径为服务器上的路径
	// 构建API请求
	apiReq := APIVoiceCloneRequest{
		ModelName:   voiceClone.ModelName,
		PromptFile:  path.Join("/data/aigc-ops", uploadPath, filepath.Base(fullPromptPath)),
		PromptText:  voiceClone.PromptText,
		SpeakerName: fmt.Sprintf("%d_%s", voiceClone.UserID, voiceClone.SpeakerName),
	}

	// 序列化请求
	reqData, err := json.Marshal(apiReq)
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "序列化请求失败: " + err.Error(),
		})
		return
	}

	// 发送克隆请求
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

	// 3. 下载音色权重文件
	speakerName := apiReq.SpeakerName
	downloadURL := fmt.Sprintf("%s/v1/file/view?key=model/tts_models/checkpoint/CosyVoice2-0.5B_1/spk_info/%s.pt", config.AppConfig.FileServerBaseURL, speakerName)

	// 创建下载目录
	voiceDir := filepath.Join(fmt.Sprint(voiceClone.UserID), config.AppConfig.VoiceDir)
	fullVoiceDir := filepath.Join(config.AppConfig.DataDir, voiceDir)

	if err := os.MkdirAll(fullVoiceDir, 0755); err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "创建下载目录失败: " + err.Error(),
		})
		return
	}

	// 下载文件
	downloadResp, err := http.Get(downloadURL)
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "下载音色权重文件失败: " + err.Error(),
		})
		return
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "下载音色权重文件失败，状态码: " + fmt.Sprint(downloadResp.StatusCode),
		})
		return
	}

	// 保存文件
	resultFile := filepath.Join(voiceDir, speakerName+".pt")
	fullResultFile := filepath.Join(config.AppConfig.DataDir, resultFile)
	out, err := os.Create(fullResultFile)
	if err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "创建音色权重文件失败: " + err.Error(),
		})
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, downloadResp.Body); err != nil {
		db.DB.Model(voiceClone).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "保存音色权重文件失败: " + err.Error(),
		})
		return
	}

	// 更新任务状态
	db.DB.Model(voiceClone).Updates(map[string]interface{}{
		"status": "completed",
		"result": resultFile,
	})

	// 自动添加到音色库
	// 创建音色库记录
	voiceLibrary := models.VoiceLibrary{
		Name:        voiceClone.SpeakerName,
		Description: voiceClone.Description,
		ModelName:   voiceClone.ModelName,
		ModelFile:   voiceClone.Result,
		Type:        "cloned",
		OwnerID:     voiceClone.UserID,
		IsPublic:    false,
	}
	result := db.DB.Create(&voiceLibrary)
	if result.Error != nil {
		log.Println("创建音色库记录失败:", result.Error)
		return
	}
}
