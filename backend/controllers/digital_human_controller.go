package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

// DigitalHumanRequest 数字人合成请求
type DigitalHumanRequest struct {
	Name            string `form:"name" binding:"required"`
	Description     string `form:"description"`
	TaskCode        string `form:"task_code" binding:"required"`
	Chaofen         int    `form:"chaofen"`
	WatermarkSwitch int    `form:"watermark_switch"`
	PN              int    `form:"pn"`
	// 音频和视频文件通过multipart/form-data上传
}

// DigitalHumanResponse 数字人合成响应
type DigitalHumanResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	AudioURL        string    `json:"audio_url"`
	VideoURL        string    `json:"video_url"`
	TaskCode        string    `json:"task_code"`
	Chaofen         int       `json:"chaofen"`
	WatermarkSwitch int       `json:"watermark_switch"`
	PN              int       `json:"pn"`
	Status          string    `json:"status"`
	ResultURL       string    `json:"result_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// APIDigitalHumanRequest API数字人合成请求
type APIDigitalHumanRequest struct {
	AudioURL        string `json:"audio_url"`
	VideoURL        string `json:"video_url"`
	Code            string `json:"code"`
	Chaofen         int    `json:"chaofen"`
	WatermarkSwitch int    `json:"watermark_switch"`
	PN              int    `json:"pn"`
}

// CreateDigitalHuman 创建数字人合成任务
func CreateDigitalHuman(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 解析表单数据
	var req DigitalHumanRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 获取上传的音频文件
	audioFile, err := c.FormFile("audio_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供音频文件"})
		return
	}

	// 检查音频文件类型
	audioExt := filepath.Ext(audioFile.Filename)
	if audioExt != ".wav" && audioExt != ".mp3" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持WAV或MP3格式的音频文件"})
		return
	}

	// 获取上传的视频文件
	videoFile, err := c.FormFile("video_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供视频文件"})
		return
	}

	// 检查视频文件类型
	videoExt := filepath.Ext(videoFile.Filename)
	if videoExt != ".mp4" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持MP4格式的视频文件"})
		return
	}

	// 生成唯一文件名并保存音频文件
	audioUniqueID := uuid.New().String()
	audioFileName := fmt.Sprintf("%s%s", audioUniqueID, audioExt)
	audioFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, audioFileName)
	fullAudioFilePath := filepath.Join(config.AppConfig.DataDir, audioFilePath)
	if err := c.SaveUploadedFile(audioFile, fullAudioFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存音频文件失败: " + err.Error()})
		return
	}

	// 上传音频文件到远程服务器
	audioBody := &bytes.Buffer{}
	audioWriter := multipart.NewWriter(audioBody)
	audioFileHandle, err := audioFile.Open()
	if err != nil {
		os.Remove(fullAudioFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开音频文件失败: " + err.Error()})
		return
	}
	defer audioFileHandle.Close()

	part, err := audioWriter.CreateFormFile("attachment", audioFileName)
	if err != nil {
		os.Remove(fullAudioFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建音频表单失败: " + err.Error()})
		return
	}
	if _, err = io.Copy(part, audioFileHandle); err != nil {
		os.Remove(fullAudioFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "复制音频文件失败: " + err.Error()})
		return
	}

	audioUploadPath := fmt.Sprintf("vhs/%d/audios", userID)
	audioWriter.WriteField("path", audioUploadPath)
	audioWriter.Close()

	audioResp, err := http.Post(config.AppConfig.FileUploadAPI, audioWriter.FormDataContentType(), audioBody)
	if err != nil {
		os.Remove(fullAudioFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传音频文件到远程服务器失败: " + err.Error()})
		return
	}
	defer audioResp.Body.Close()

	// 生成唯一文件名并保存视频文件
	videoUniqueID := uuid.New().String()
	videoFileName := fmt.Sprintf("%s%s", videoUniqueID, videoExt)
	videoFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.VideoDir, videoFileName)
	fullVideoFilePath := filepath.Join(config.AppConfig.DataDir, videoFilePath)
	if err := c.SaveUploadedFile(videoFile, fullVideoFilePath); err != nil {
		// 删除已上传的音频文件
		os.Remove(fullAudioFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存视频文件失败: " + err.Error()})
		return
	}

	// 上传视频文件到远程服务器
	videoBody := &bytes.Buffer{}
	videoWriter := multipart.NewWriter(videoBody)
	videoFileHandle, err := videoFile.Open()
	if err != nil {
		os.Remove(fullAudioFilePath)
		os.Remove(fullVideoFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开视频文件失败: " + err.Error()})
		return
	}
	defer videoFileHandle.Close()

	part, err = videoWriter.CreateFormFile("attachment", videoFileName)
	if err != nil {
		os.Remove(fullAudioFilePath)
		os.Remove(fullVideoFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建视频表单失败: " + err.Error()})
		return
	}
	if _, err = io.Copy(part, videoFileHandle); err != nil {
		os.Remove(fullAudioFilePath)
		os.Remove(fullVideoFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "复制视频文件失败: " + err.Error()})
		return
	}

	videoUploadPath := fmt.Sprintf("vhs/%d/videos", userID)
	videoWriter.WriteField("path", videoUploadPath)
	videoWriter.Close()

	videoResp, err := http.Post(config.AppConfig.FileUploadAPI, videoWriter.FormDataContentType(), videoBody)
	if err != nil {
		os.Remove(fullAudioFilePath)
		os.Remove(fullVideoFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传视频文件到远程服务器失败: " + err.Error()})
		return
	}
	defer videoResp.Body.Close()

	// 创建数字人合成记录
	digitalHuman := models.DigitalHuman{
		UserID:          userID.(uint),
		Name:            req.Name,
		Description:     req.Description,
		AudioURL:        audioFilePath,
		VideoURL:        videoFilePath,
		TaskCode:        req.TaskCode,
		Chaofen:         req.Chaofen,
		WatermarkSwitch: req.WatermarkSwitch,
		PN:              req.PN,
		Status:          "pending",
	}

	result := db.DB.Create(&digitalHuman)
	if result.Error != nil {
		// 删除已上传的文件
		os.Remove(fullAudioFilePath)
		os.Remove(fullVideoFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建数字人合成记录失败: " + result.Error.Error()})
		return
	}

	// 异步调用数字人合成API
	go func() {
		// 更新状态为处理中
		db.DB.Model(&digitalHuman).Update("status", "processing")

		// 构建API请求
		apiReq := APIDigitalHumanRequest{
			AudioURL:        filepath.Join("/code/data/", audioUploadPath, audioFileName),
			VideoURL:        filepath.Join("/code/data/", videoUploadPath, videoFileName),
			Code:            req.TaskCode,
			Chaofen:         req.Chaofen,
			WatermarkSwitch: req.WatermarkSwitch,
			PN:              req.PN,
		}

		// 序列化请求
		reqData, err := json.Marshal(apiReq)
		if err != nil {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "序列化请求失败: " + err.Error(),
			})
			return
		}

		// 发送请求
		client := &http.Client{}
		req, err := http.NewRequest("POST", config.AppConfig.DigitalHumanAPI, bytes.NewBuffer(reqData))
		if err != nil {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "创建请求失败: " + err.Error(),
			})
			return
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "调用API失败: " + err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "读取API响应失败: " + err.Error(),
			})
			return
		}

		// 解析响应
		var apiResp map[string]interface{}
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "解析API响应失败: " + err.Error(),
			})
			return
		}

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": fmt.Sprintf("API返回错误: %d %s", resp.StatusCode, string(respBody)),
			})
			return
		}

		// 更新任务状态为进行中，等待后续查询结果
		db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
			"status": "processing",
		})
	}()

	// 返回响应
	c.JSON(http.StatusCreated, gin.H{
		"message": "数字人合成任务已创建",
		"digital_human": DigitalHumanResponse{
			ID:              digitalHuman.ID,
			Name:            digitalHuman.Name,
			Description:     digitalHuman.Description,
			AudioURL:        utils.GetFileURL(digitalHuman.AudioURL),
			VideoURL:        utils.GetFileURL(digitalHuman.VideoURL),
			TaskCode:        digitalHuman.TaskCode,
			Chaofen:         digitalHuman.Chaofen,
			WatermarkSwitch: digitalHuman.WatermarkSwitch,
			PN:              digitalHuman.PN,
			Status:          digitalHuman.Status,
			ResultURL:       digitalHuman.ResultURL,
			CreatedAt:       digitalHuman.CreatedAt,
			UpdatedAt:       digitalHuman.UpdatedAt,
		},
	})
}

// QueryDigitalHumanProgress 查询数字人合成进度
func QueryDigitalHumanProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var digitalHuman models.DigitalHuman
	result := db.DB.First(&digitalHuman, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "数字人合成任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if digitalHuman.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此任务"})
		return
	}

	// 如果任务已完成或失败，直接返回状态
	if digitalHuman.Status == "completed" || digitalHuman.Status == "failed" {
		c.JSON(http.StatusOK, gin.H{
			"digital_human": DigitalHumanResponse{
				ID:              digitalHuman.ID,
				Name:            digitalHuman.Name,
				Description:     digitalHuman.Description,
				AudioURL:        utils.GetFileURL(digitalHuman.AudioURL),
				VideoURL:        utils.GetFileURL(digitalHuman.VideoURL),
				TaskCode:        digitalHuman.TaskCode,
				Chaofen:         digitalHuman.Chaofen,
				WatermarkSwitch: digitalHuman.WatermarkSwitch,
				PN:              digitalHuman.PN,
				Status:          digitalHuman.Status,
				ResultURL:       utils.GetFileURL(digitalHuman.ResultURL),
				CreatedAt:       digitalHuman.CreatedAt,
				UpdatedAt:       digitalHuman.UpdatedAt,
			},
		})
		return
	}

	// 构建查询URL
	queryURL := fmt.Sprintf("%s?code=%s", config.AppConfig.DigitalHumanQuery, digitalHuman.TaskCode)

	// 创建请求
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败: " + err.Error()})
		return
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询进度失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取响应失败: " + err.Error()})
		return
	}

	// 解析响应
	type QueryResponse struct {
		Code    int    `json:"code"`
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		Data    struct {
			Code     string  `json:"code"`
			Msg      string  `json:"msg"`
			Progress float64 `json:"progress"`
			Result   string  `json:"result"`
			Status   int     `json:"status"`
		} `json:"data"`
	}

	var apiResp QueryResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析响应失败: " + err.Error()})
		return
	}

	// 检查响应状态
	if !apiResp.Success || apiResp.Code != 10000 {
		db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": fmt.Sprintf("API返回错误: %d %s", resp.StatusCode, string(respBody)),
		})
		return
	}

	// 检查是否完成
	if apiResp.Data.Progress >= 100 {
		// 任务完成，构建结果URL
		downloadURL := fmt.Sprintf("%s/v1/file/view?key=temp%s", config.AppConfig.FileServerBaseURL, apiResp.Data.Result)
		videoDir := filepath.Join(fmt.Sprint(userID), config.AppConfig.VideoDir)
		fullVideoDir := filepath.Join(config.AppConfig.DataDir, videoDir)

		if err := os.MkdirAll(fullVideoDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败: " + err.Error()})
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "创建目录失败: " + err.Error(),
			})

			return
		}

		downloadResp, err := http.Get(downloadURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "下载文件失败: " + err.Error()})
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "下载文件失败: " + err.Error(),
			})
			return
		}
		defer downloadResp.Body.Close()

		if downloadResp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "下载文件失败，状态码: " + downloadResp.Status})
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "下载文件失败，状态码: " + downloadResp.Status,
			})
			return
		}

		// 保存文件
		videoFile := filepath.Join(videoDir, filepath.Base(apiResp.Data.Result))
		fullVideoFile := filepath.Join(config.AppConfig.DataDir, videoFile)
		out, err := os.Create(fullVideoFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件失败: " + err.Error()})
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "创建文件失败: " + err.Error(),
			})
			return
		}
		defer out.Close()
		_, err = io.Copy(out, downloadResp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "保存文件失败: " + err.Error(),
			})
			return
		}

		// 更新任务状态
		db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
			"status":     "completed",
			"result_url": videoFile,
		})

		digitalHuman.Status = "completed"
		digitalHuman.ResultURL = videoFile
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"progress": apiResp.Data.Progress,
		"digital_human": DigitalHumanResponse{
			ID:              digitalHuman.ID,
			Name:            digitalHuman.Name,
			Description:     digitalHuman.Description,
			AudioURL:        utils.GetFileURL(digitalHuman.AudioURL),
			VideoURL:        utils.GetFileURL(digitalHuman.VideoURL),
			TaskCode:        digitalHuman.TaskCode,
			Chaofen:         digitalHuman.Chaofen,
			WatermarkSwitch: digitalHuman.WatermarkSwitch,
			PN:              digitalHuman.PN,
			Status:          digitalHuman.Status,
			ResultURL:       utils.GetFileURL(digitalHuman.ResultURL),
			CreatedAt:       digitalHuman.CreatedAt,
			UpdatedAt:       digitalHuman.UpdatedAt,
		},
	})
}

// GetDigitalHuman 获取数字人合成任务
func GetDigitalHuman(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var digitalHuman models.DigitalHuman
	result := db.DB.First(&digitalHuman, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "数字人合成任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if digitalHuman.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此任务"})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"digital_human": DigitalHumanResponse{
			ID:              digitalHuman.ID,
			Name:            digitalHuman.Name,
			Description:     digitalHuman.Description,
			AudioURL:        utils.GetFileURL(digitalHuman.AudioURL),
			VideoURL:        utils.GetFileURL(digitalHuman.VideoURL),
			TaskCode:        digitalHuman.TaskCode,
			Chaofen:         digitalHuman.Chaofen,
			WatermarkSwitch: digitalHuman.WatermarkSwitch,
			PN:              digitalHuman.PN,
			Status:          digitalHuman.Status,
			ResultURL:       utils.GetFileURL(digitalHuman.ResultURL),
			CreatedAt:       digitalHuman.CreatedAt,
			UpdatedAt:       digitalHuman.UpdatedAt,
		},
	})
}

// ListDigitalHumans 获取数字人合成任务列表
func ListDigitalHumans(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 查询任务总数
	var count int64
	db.DB.Model(&models.DigitalHuman{}).Where("user_id = ?", userID).Count(&count)

	// 查询任务列表
	var digitalHumans []models.DigitalHuman
	result := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&digitalHumans)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 构建响应
	responses := make([]DigitalHumanResponse, len(digitalHumans))
	for i, dh := range digitalHumans {
		responses[i] = DigitalHumanResponse{
			ID:              dh.ID,
			Name:            dh.Name,
			Description:     dh.Description,
			AudioURL:        utils.GetFileURL(dh.AudioURL),
			VideoURL:        utils.GetFileURL(dh.VideoURL),
			TaskCode:        dh.TaskCode,
			Chaofen:         dh.Chaofen,
			WatermarkSwitch: dh.WatermarkSwitch,
			PN:              dh.PN,
			Status:          dh.Status,
			ResultURL:       utils.GetFileURL(dh.ResultURL),
			CreatedAt:       dh.CreatedAt,
			UpdatedAt:       dh.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":          count,
		"page":           page,
		"size":           size,
		"digital_humans": responses,
	})
}

// DeleteDigitalHuman 删除数字人合成任务
func DeleteDigitalHuman(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务
	var digitalHuman models.DigitalHuman
	result := db.DB.First(&digitalHuman, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "数字人合成任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		}
		return
	}

	// 检查权限
	if digitalHuman.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此任务"})
		return
	}

	// 删除关联文件
	if digitalHuman.AudioURL != "" {
		os.Remove(filepath.Join(config.AppConfig.DataDir, digitalHuman.AudioURL))
	}
	if digitalHuman.VideoURL != "" {
		os.Remove(filepath.Join(config.AppConfig.DataDir, digitalHuman.VideoURL))
	}
	if digitalHuman.ResultURL != "" {
		os.Remove(filepath.Join(config.AppConfig.DataDir, digitalHuman.ResultURL))
	}

	// 删除记录
	result = db.DB.Delete(&digitalHuman)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "数字人合成任务已删除"})
}
