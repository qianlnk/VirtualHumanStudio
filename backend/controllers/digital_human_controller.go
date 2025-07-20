package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/middleware"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/services"
	"github.com/qianlnk/VirtualHumanStudio/backend/storages"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// DigitalHumanRequest 数字人合成请求
type DigitalHumanRequest struct {
	Name            string `form:"name" binding:"required"`
	Description     string `form:"description"`
	TaskCode        string `form:"task_code"`
	Chaofen         int    `form:"chaofen"`
	WatermarkSwitch int    `form:"watermark_switch"`
	PN              int    `form:"pn"`
	TTSTaskID       string `form:"tts_task_id"`
	// 音频和视频文件通过multipart/form-data上传
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

func InitDigitalHumanConsumer() {
	services.StartDigitalHumanQueueConsumer(context.Background(), processDigitalHumanTask)
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

	// 如果task_code为空，则生成一个随机任务码
	if req.TaskCode == "" {
		req.TaskCode = uuid.New().String()
	}

	var audio io.Reader
	var audioExt string
	// 是否从TTS获取音频
	if req.TTSTaskID != "" {
		// 获取TTS任务
		ttsTask := &models.TTSTask{}
		result := db.DB.First(ttsTask, req.TTSTaskID)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TTS任务不存在"})
			return
		}
		// 获取TTS任务的音频文件
		audioFile, err := storages.Client.OpenFile(context.Background(), ttsTask.OutputFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TTS任务的音频文件不存在"})
			return
		}
		audioExt = filepath.Ext(ttsTask.OutputFile)
		audio = audioFile
	} else {
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

		tmpFile, err := audioFile.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "打开音频文件失败: " + err.Error()})
			return
		}
		audio = tmpFile
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

	err = storages.Client.SaveFile(context.Background(), audioFilePath, audio)
	if err != nil {
		log.Println("保存音频文件失败: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存音频文件失败: " + err.Error()})
		return
	}

	// 生成唯一文件名并保存视频文件
	videoUniqueID := uuid.New().String()
	videoFileName := fmt.Sprintf("%s%s", videoUniqueID, videoExt)
	videoFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.VideoDir, videoFileName)

	video, err := videoFile.Open()
	if err != nil {
		log.Println("打开视频文件失败: ", err)
		storages.Client.Delete(context.Background(), audioFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开视频文件失败: " + err.Error()})
		return
	}
	err = storages.Client.SaveFile(context.Background(), videoFilePath, video)
	if err != nil {
		log.Println("保存视频文件失败: ", err)
		storages.Client.Delete(context.Background(), audioFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存视频文件失败: " + err.Error()})
		return
	}

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
		storages.Client.Delete(context.Background(), digitalHuman.AudioURL)
		storages.Client.Delete(context.Background(), digitalHuman.VideoURL)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建数字人合成记录失败: " + result.Error.Error()})
		return
	}

	// 添加到任务队列
	isMember := middleware.IsMember(digitalHuman.UserID)
	err = services.AddToDigitalHumanQueue(context.Background(), digitalHuman.ID, digitalHuman.UserID, isMember)
	if err != nil {
		db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "添加到队列失败: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加到数字人合成队列失败"})
		return
	}

	digitalHuman.AudioURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.AudioURL)
	digitalHuman.VideoURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.VideoURL)

	// 返回响应
	// 返回响应
	c.JSON(http.StatusCreated, gin.H{
		"message":       "数字人合成任务已创建",
		"digital_human": digitalHuman,
	})
}

func processDigitalHumanTask(ctx context.Context, taskID uint) error {
	// 查询任务
	var digitalHuman = &models.DigitalHuman{}
	result := db.DB.First(digitalHuman, taskID)
	if result.Error != nil {
		return fmt.Errorf("查询任务失败: %v", result.Error)
	}

	// 更新任务状态为进行中
	log.Println("更新任务:", digitalHuman.Name, "状态为进行中")
	result = db.DB.Model(digitalHuman).Update("status", "processing")
	if result.Error != nil {
		return fmt.Errorf("更新任务状态失败: %v", result.Error)
	}

	// 上传音频文件到远程服务器
	log.Println("上传音频文件到远程服务器:", digitalHuman.AudioURL)
	audioFileHandle, err := storages.Client.OpenFile(context.Background(), digitalHuman.AudioURL)
	if err != nil {
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "获取音频文件失败: " + err.Error(),
		})
		return fmt.Errorf("获取音频文件失败: %v", err)
	}
	defer audioFileHandle.Close()

	audioFileName := filepath.Base(digitalHuman.AudioURL)
	audioBody := &bytes.Buffer{}
	audioWriter := multipart.NewWriter(audioBody)

	part, err := audioWriter.CreateFormFile("attachment", audioFileName)
	if err != nil {
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "创建音频表单失败: " + err.Error(),
		})
		return err
	}
	if _, err = io.Copy(part, audioFileHandle); err != nil {
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "复制音频文件失败: " + err.Error(),
		})
		return err
	}

	audioUploadPath := fmt.Sprintf("vhs/%d/audios", digitalHuman.UserID)
	audioWriter.WriteField("path", audioUploadPath)
	audioWriter.Close()

	audioResp, err := http.Post(config.AppConfig.FileUploadAPI, audioWriter.FormDataContentType(), audioBody)
	if err != nil {
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "上传音频文件到远程服务器失败: " + err.Error(),
		})
		return err
	}
	defer audioResp.Body.Close()

	// 上传视频文件到远程服务器
	log.Println("上传视频文件到远程服务器:", digitalHuman.VideoURL)
	videoFileHandle, err := storages.Client.OpenFile(context.Background(), digitalHuman.VideoURL)
	if err != nil {
		log.Println("获取视频文件失败:", err)
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "获取视频文件失败: " + err.Error(),
		})
		return err
	}
	defer videoFileHandle.Close()
	videoFileName := filepath.Base(digitalHuman.VideoURL)
	videoBody := &bytes.Buffer{}
	videoWriter := multipart.NewWriter(videoBody)
	log.Println("创建视频表单")
	part, err = videoWriter.CreateFormFile("attachment", videoFileName)
	if err != nil {
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "创建视频表单失败: " + err.Error(),
		})
		return err
	}
	log.Println("复制视频文件")
	if _, err = io.Copy(part, videoFileHandle); err != nil {
		log.Println("复制视频文件失败:", err)
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "复制视频文件失败: " + err.Error(),
		})
		return err
	}

	videoUploadPath := fmt.Sprintf("vhs/%d/videos", digitalHuman.UserID)
	videoWriter.WriteField("path", videoUploadPath)
	videoWriter.Close()

	log.Println("上传视频文件到远程服务器:", videoUploadPath)
	videoResp, err := http.Post(config.AppConfig.FileUploadAPI, videoWriter.FormDataContentType(), videoBody)
	if err != nil {
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "上传视频文件到远程服务器失败: " + err.Error(),
		})
		return err
	}
	defer videoResp.Body.Close()

	// 构建API请求
	apiReq := APIDigitalHumanRequest{
		AudioURL:        filepath.Join("/code/data/", audioUploadPath, audioFileName),
		VideoURL:        filepath.Join("/code/data/", videoUploadPath, videoFileName),
		Code:            digitalHuman.TaskCode,
		Chaofen:         digitalHuman.Chaofen,
		WatermarkSwitch: digitalHuman.WatermarkSwitch,
		PN:              digitalHuman.PN,
	}

	// 序列化请求
	reqData, err := json.Marshal(apiReq)
	if err != nil {
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "序列化请求失败: " + err.Error(),
		})
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送请求
	log.Println("开始合成数字人")
	client := &http.Client{}
	req, err := http.NewRequest("POST", config.AppConfig.DigitalHumanAPI, bytes.NewBuffer(reqData))
	if err != nil {
		log.Println("创建请求失败:", err)
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "创建请求失败: " + err.Error(),
		})
		return fmt.Errorf("创建请求失败: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("调用API失败:", err)
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "调用API失败: " + err.Error(),
		})
		return fmt.Errorf("调用API失败: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取API响应失败:", err)
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "读取API响应失败: " + err.Error(),
		})
		return fmt.Errorf("读取API响应失败: %v", err)
	}
	// 解析响应
	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		log.Println("解析API响应失败:", err)
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "解析API响应失败: " + err.Error(),
		})
		return fmt.Errorf("解析API响应失败: %v", err)
	}
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		log.Println("API返回错误:", resp.StatusCode, string(respBody))
		db.DB.Model(digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": fmt.Sprintf("API返回错误: %d %s", resp.StatusCode, string(respBody)),
		})
		return fmt.Errorf("API返回错误: %d %s", resp.StatusCode, string(respBody))
	}

	// 轮询查询任务状态
	for {
		log.Println("查询任务进度")
		progress, err := queryProgress(digitalHuman, true)
		if err != nil {
			db.DB.Model(digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "查询任务进度失败: " + err.Error(),
			})
			return fmt.Errorf("查询任务进度失败: %v", err)
		}

		if progress >= 100 {
			break
		}
		time.Sleep(time.Second * 5)
	}

	return nil
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
	var digitalHuman = &models.DigitalHuman{}
	result := db.DB.First(digitalHuman, id)
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

	var progress int64
	if digitalHuman.Status == "processing" {
		progress = int64(time.Since(digitalHuman.CreatedAt).Seconds() / (60 * 5) * 100)

		if progress > 100 {
			progress = 99
		}
	}

	digitalHuman.AudioURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.AudioURL)
	digitalHuman.VideoURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.VideoURL)
	digitalHuman.ResultURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.ResultURL)

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"progress":      progress,
		"digital_human": digitalHuman,
	})
}

func queryProgress(digitalHuman *models.DigitalHuman, save bool) (float64, error) {
	// 构建查询URL
	queryURL := fmt.Sprintf("%s?code=%s", config.AppConfig.DigitalHumanQuery, digitalHuman.TaskCode)
	fmt.Println("正在查询进度：", queryURL)
	// 创建请求
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return 0, err
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	// 检查响应状态
	if !apiResp.Success {
		db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": fmt.Sprintf("API返回错误: %d %s", resp.StatusCode, string(respBody)),
		})

		return 0, err
	}

	// 检查是否完成
	if apiResp.Data.Progress >= 100 && save {
		// 任务完成，构建结果URL
		downloadURL := fmt.Sprintf("%s/v1/file/view?key=temp%s", config.AppConfig.FileServerBaseURL, apiResp.Data.Result)
		videoDir := filepath.Join(fmt.Sprint(digitalHuman.UserID), config.AppConfig.VideoDir)

		// 下载文件
		fmt.Println("正在下载文件：", downloadURL)

		downloadResp, err := http.Get(downloadURL)
		if err != nil {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "下载文件失败: " + err.Error(),
			})
			return 0, err
		}
		defer downloadResp.Body.Close()

		if downloadResp.StatusCode != http.StatusOK {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "下载文件失败，状态码: " + downloadResp.Status,
			})
			return 0, err
		}

		// 保存文件
		videoFile := filepath.Join(videoDir, filepath.Base(apiResp.Data.Result))
		err = storages.Client.SaveFile(context.Background(), videoFile, downloadResp.Body)
		if err != nil {
			db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "保存文件失败: " + err.Error(),
			})
			return 0, err
		}

		// 更新任务状态
		db.DB.Model(&digitalHuman).Updates(map[string]interface{}{
			"status":     "completed",
			"result_url": videoFile,
		})

		digitalHuman.Status = "completed"
		digitalHuman.ResultURL = videoFile
	}

	return apiResp.Data.Progress, nil
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

	digitalHuman.AudioURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.AudioURL)
	digitalHuman.VideoURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.VideoURL)
	digitalHuman.ResultURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHuman.ResultURL)

	position, queueType, _ := services.GetDigitalHumanQueuePosition(context.Background(), digitalHuman.ID)
	digitalHuman.QueuePosition = position
	digitalHuman.QueueType = queueType

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"digital_human": digitalHuman,
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
	for i := range digitalHumans {
		digitalHumans[i].AudioURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHumans[i].AudioURL)
		digitalHumans[i].VideoURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHumans[i].VideoURL)
		digitalHumans[i].ResultURL, _ = storages.Client.GetFileUrl(context.Background(), digitalHumans[i].ResultURL)

		if digitalHumans[i].Status == "pending" {
			position, queueType, _ := services.GetDigitalHumanQueuePosition(context.Background(), digitalHumans[i].ID)
			digitalHumans[i].QueuePosition = position
			digitalHumans[i].QueueType = queueType
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":          count,
		"page":           page,
		"size":           size,
		"digital_humans": digitalHumans,
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
		storages.Client.Delete(context.Background(), digitalHuman.AudioURL)
	}
	if digitalHuman.VideoURL != "" {
		storages.Client.Delete(context.Background(), digitalHuman.VideoURL)
	}
	if digitalHuman.ResultURL != "" {
		storages.Client.Delete(context.Background(), digitalHuman.ResultURL)
	}

	// 删除记录
	result = db.DB.Delete(&digitalHuman)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "数字人合成任务已删除"})
}
