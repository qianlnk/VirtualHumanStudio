package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/storages"

	"github.com/gin-gonic/gin"
)

// 从上下文中获取用户ID
func getUserIDFromContext(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	// 根据类型进行转换
	switch id := userID.(type) {
	case uint:
		return strconv.FormatUint(uint64(id), 10)
	case int:
		return strconv.Itoa(id)
	case string:
		return id
	default:
		// 尝试使用fmt转为字符串
		return fmt.Sprintf("%v", userID)
	}
}

// 构建私有路径
func buildPrivatePath(userID, path string) string {
	return "vhs/" + userID + "/" + path
}

// 构建公共路径
func buildPublicPath(userID, path string) string {
	return "vhs/public/" + userID + "/" + path
}

// 替换域名
func replaceDomain(urlStr string) string {
	// 如果配置了自定义域名，则替换
	if config.AppConfig.Domain != "" {
		u, err := url.Parse(urlStr)
		if err != nil {
			log.Printf("parse url error: %v", err)
			return urlStr
		}

		// 替换域名
		oldHost := u.Host
		u.Host = config.AppConfig.Domain
		log.Printf("replace domain from %s to %s", oldHost, u.Host)

		return u.String()
	}
	return urlStr
}

// GetUploadURL 获取上传URL的处理函数
func GetUploadURL(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	privatePath := buildPrivatePath(userID, req.Path)
	log.Printf("privatePath: %s", privatePath)

	policy := &storages.PutPolicy{
		Key:     privatePath,
		Expires: 1800,
	}

	method, host, _, _, _, err := storages.UploadClient.UploadToken(c.Request.Context(), policy)
	if err != nil {
		log.Printf("get upload token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取上传凭证失败"})
		return
	}

	// 获取访问URL
	visitURL, err := storages.UploadClient.GetFileUrl(c.Request.Context(), privatePath)
	if err != nil {
		log.Printf("get file url error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取访问链接失败"})
		return
	}

	// visitURL = replaceDomain(visitURL)

	c.JSON(http.StatusOK, gin.H{
		"method":     method,
		"upload_url": host,
		"visit_url":  visitURL,
	})
}

// GetVisitURL 获取访问URL的处理函数
func GetVisitURL(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	privatePath := buildPrivatePath(userID, req.Path)

	url, err := storages.UploadClient.GetFileUrl(c.Request.Context(), privatePath)
	if err != nil {
		log.Printf("get file url error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取访问链接失败"})
		return
	}

	url = replaceDomain(url)
	log.Printf("privatePath: %s, url: %s", privatePath, url)

	c.JSON(http.StatusOK, gin.H{
		"visit_url": url,
		"expire_at": time.Now().Add(time.Second * 300).Unix(),
	})
}

// GetPublicUploadURL 获取公共上传URL的处理函数
func GetPublicUploadURL(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	publicPath := buildPublicPath(userID, req.Path)
	log.Printf("publicPath: %s", publicPath)

	policy := &storages.PutPolicy{
		Key:     publicPath,
		Expires: 1800,
	}

	method, host, _, _, _, err := storages.UploadClient.UploadToken(c.Request.Context(), policy)
	if err != nil {
		log.Printf("get upload token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取上传凭证失败"})
		return
	}

	uri, err := url.Parse(host)
	if err != nil {
		log.Printf("parse url error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析URL失败"})
		return
	}

	visitURL := uri.Scheme + "://" + uri.Host + uri.Path
	// visitURL = replaceDomain(visitURL)

	c.JSON(http.StatusOK, gin.H{
		"method":     method,
		"upload_url": host,
		"visit_url":  visitURL,
	})
}

// GetPublicVisitURL 获取公共访问URL的处理函数
func GetPublicVisitURL(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	publicPath := buildPublicPath(userID, req.Path)

	fileURL, err := storages.UploadClient.GetFileUrl(c.Request.Context(), publicPath)
	if err != nil {
		log.Printf("get file url error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取访问链接失败"})
		return
	}

	uri, err := url.Parse(fileURL)
	if err != nil {
		log.Printf("parse url error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析URL失败"})
		return
	}

	fileURL = uri.Scheme + "://" + uri.Host + uri.Path
	// fileURL = replaceDomain(fileURL)

	log.Printf("publicPath: %s, url: %s", publicPath, fileURL)

	c.JSON(http.StatusOK, gin.H{
		"visit_url": fileURL,
		"expire_at": time.Now().Add(time.Second * 300).Unix(),
	})
}

// 文件查看处理函数
func FileView(c *gin.Context) {
	// 获取文件路径参数
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供文件路径"})
		return
	}
	// // 获取用户ID
	// userID, exists := c.Get("user_id")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
	// 	return
	// }

	// 构建完整文件路径
	fullPath := filepath.Join(config.AppConfig.DataDir, filePath)
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// // 检查文件访问权限
	// if !checkFileAccess(uint(userID.(uint)), filePath) {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该文件"})
	// 	return
	// }

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

	// 检查文件是否属于asr任务
	var asrTask models.ASRTask
	result = db.DB.Where("user_id =? AND input_file =?", userID, filePath).First(&asrTask)
	if result.Error == nil {
		return true
	}

	// 检查文件是否属于饰品替换任务
	var accessory models.Accessory
	result = db.DB.Where("user_id =? AND (item_image =? OR model_image =? OR mask_image =? OR result_image=?)", userID, filePath, filePath, filePath, filePath).First(&accessory)
	if result.Error == nil {
		return true
	}

	// 检查文件是否属于通用图片处理任务
	var imageProcessingTask models.ComfyUIWorkflowTask
	likeFilePath := "%" + filePath + "%"
	result = db.DB.Where("user_id =? AND (input_params like ? OR output_params like ?)", userID, likeFilePath, likeFilePath).First(&imageProcessingTask)
	if result.Error == nil {
		return true
	}

	return false
}
