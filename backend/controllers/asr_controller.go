package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/middleware"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/services"
	"github.com/qianlnk/VirtualHumanStudio/backend/storages"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ASRRequest ASR请求参数
type ASRRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	Model       string `form:"model"`
	AudioURL    string `form:"audio_url"`
}

// ASRSentence 语音识别句子结果
type ASRSentence struct {
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// ASRResponse ASR响应结果
type ASRResponse struct {
	Text          string           `json:"text"`
	Sentences     []ASRSentence    `json:"sentences"`
	Code          int              `json:"code"`
	TimeCostMs    map[string]int64 `json:"time_cost"`
	QueuePosition int              `json:"queue_position"`
	QueueType     string           `json:"queue_type"`
	models.ASRTask
}

func InitASRQueueConsumer() {
	services.StartASRQueueConsumer(context.Background(), processASRTask)
}

// saveAudioFile 保存音频文件到本地
func saveAudioFile(userID uint, fileName string, reader io.Reader) (string, error) {
	ext := filepath.Ext(fileName)
	if ext != ".wav" && ext != ".mp3" {
		return "", fmt.Errorf("仅支持WAV或MP3格式的音频文件")
	}

	// 生成唯一文件名
	newFileName := uuid.New().String() + ext
	filePath := utils.GetUserFilePath(userID, config.AppConfig.UploadDir, newFileName)

	err := storages.Client.SaveFile(context.Background(), filePath, reader)
	if err != nil {
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	return filePath, nil
}

// callASRService 调用ASR服务
func callASRService(model string, audioData io.Reader, audioFileName string) (*ASRResponse, error) {
	// 准备ASR请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("model", model)

	// 添加音频文件
	part, err := writer.CreateFormFile("audio", audioFileName)
	if err != nil {
		return nil, fmt.Errorf("创建表单文件失败: %v", err)
	}

	if _, err = io.Copy(part, audioData); err != nil {
		return nil, fmt.Errorf("写入音频数据失败: %v", err)
	}
	writer.Close()

	// 调用ASR服务
	fmt.Println(config.AppConfig.ASRAPI)
	resp, err := http.Post(config.AppConfig.ASRAPI, writer.FormDataContentType(), body)
	if err != nil {
		return nil, fmt.Errorf("调用ASR服务失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ASR服务响应错误: %s", resp.Status)
	}

	// 解析ASR服务响应
	rsp := &ASRResponse{
		TimeCostMs: make(map[string]int64),
	}
	if err := utils.DecodeJSON(resp.Body, rsp); err != nil {
		return nil, fmt.Errorf("解析ASR服务响应失败: %v", err)
	}

	return rsp, nil
}

// ASR 语音识别接口
func CreateASRTask(c *gin.Context) {
	logger := logrus.WithContext(c)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 解析请求参数
	var req ASRRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Errorf("Invalid request parameters: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	req.Model = "llm-asr_1"

	// 创建ASR任务
	asrTask := models.ASRTask{
		UserID:      userID.(uint),
		Name:        req.Name,
		Description: req.Description,
		Model:       req.Model,
		Status:      "pending",
	}

	var audioReader io.Reader
	var audioFileName string

	// 处理音频来源
	if req.AudioURL == "" {
		// 处理上传的音频文件
		file, err := c.FormFile("audio_file")
		if err != nil {
			logger.Error("No audio URL and no uploaded file")
			c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供音频URL或上传音频文件"})
			return
		}

		src, err := file.Open()
		if err != nil {
			logger.Errorf("Error opening uploaded file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打开上传文件失败: " + err.Error()})
			return
		}
		defer src.Close()

		audioReader = src
		audioFileName = file.Filename
	} else {
		// 处理音频URL
		u, err := url.Parse(req.AudioURL)
		if err != nil {
			logger.Errorf("Error parsing audio URL %s: %v", req.AudioURL, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "解析音频URL失败: " + err.Error()})
			return
		}

		resp, err := http.Get(req.AudioURL)
		if err != nil {
			logger.Errorf("Error downloading audio from %s: %v", req.AudioURL, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "下载音频文件失败: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Errorf("Error downloading audio, status: %s", resp.Status)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "下载音频文件失败: " + resp.Status})
			return
		}

		audioReader = resp.Body
		audioFileName = filepath.Base(u.Path)
	}

	// 保存音频文件
	filePath, err := saveAudioFile(userID.(uint), audioFileName, audioReader)
	if err != nil {
		logger.Errorf("Error saving audio file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	asrTask.InputFile = filePath

	// 保存任务记录
	if err := db.DB.Create(&asrTask).Error; err != nil {
		logger.Errorf("Error creating ASR task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	// 添加任务到队列
	isMember := middleware.IsMember(asrTask.UserID)
	err = services.AddToASRQueue(context.Background(), asrTask.ID, asrTask.UserID, isMember)
	if err != nil {
		logger.Errorf("Error adding task to queue: %v", err)
		db.DB.Model(&asrTask).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "添加到队列失败: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加任务到队列失败"})
		return
	}

	position, queueType, err := services.GetASRQueuePosition(context.Background(), asrTask.ID)
	if err != nil {
		logger.Errorf("Error getting queue position: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ASR任务已创建",
		"asr_task": ASRResponse{
			ASRTask:       asrTask,
			QueuePosition: position,
			QueueType:     queueType,
		},
	})
}

func processASRTask(ctx context.Context, taskID uint) error {
	// 查询任务
	var task models.ASRTask
	result := db.DB.First(&task, taskID)
	if result.Error != nil {
		return fmt.Errorf("查询任务失败: %v", result.Error)
	}
	// 更新状态为处理中
	db.DB.Model(&task).Update("status", "processing")

	logrus.Infof("Processing ASR task: %s", utils.ToJSONString(task))

	// 调用ASR服务
	audioFile, err := storages.Client.OpenFile(context.Background(), task.InputFile)
	if err != nil {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "打开音频文件失败: " + err.Error(),
		})
		return fmt.Errorf("打开音频文件失败: %v", err)
	}
	defer audioFile.Close()
	rsp, err := callASRService(task.Model, audioFile, filepath.Base(task.InputFile))
	if err != nil {
		db.DB.Model(&task).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "调用ASR服务失败: " + err.Error(),
		})
		return fmt.Errorf("调用ASR服务失败: %v", err)
	}
	// 处理ASR响应
	if rsp.Code != 0 {
		if rsp.Code == 1 && strings.Contains(rsp.Text, "音频中没有有效语音") {
			task.Status = "completed"
			task.OutputText = "音频中没有有效语音"
		} else {
			task.Status = "failed"
			task.ErrorMsg = rsp.Text
			return fmt.Errorf("ASR服务错误: %s", rsp.Text)
		}
	} else {
		task.Status = "completed"
		task.OutputText = rsp.Text
	}

	// 更新任务状态
	result = db.DB.Save(&task)
	if result.Error != nil {
		return fmt.Errorf("更新任务状态失败: %v", result.Error)
	}

	return nil
}

// GetASRTask 获取ASR任务详情
func GetASRTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	taskID := c.Param("id")

	// 查询任务
	var task models.ASRTask
	result := db.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	var err error
	task.InputFile, err = storages.Client.GetFileUrl(context.Background(), task.InputFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件URL失败"})
		return
	}

	// 任务队列位置
	position, queueType, err := services.GetASRQueuePosition(context.Background(), task.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务队列位置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": ASRResponse{
			ASRTask:       task,
			QueuePosition: position,
			QueueType:     queueType,
		},
	})
}

// ListASRTasks 获取ASR任务列表
func ListASRTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 查询总数
	var count int64
	if err := db.DB.Model(&models.ASRTask{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 查询列表
	var tasks []*models.ASRTask
	if err := db.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	asrRes := make([]*ASRResponse, len(tasks))

	for i, task := range tasks {
		var err error
		task.InputFile, err = storages.Client.GetFileUrl(context.Background(), task.InputFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件URL失败"})
			return
		}

		asrRes[i] = &ASRResponse{
			ASRTask: *task,
		}
		// 任务队列位置
		if task.Status == "pending" {
			position, queueType, err := services.GetASRQueuePosition(context.Background(), task.ID)
			if err != nil {
				continue
			}

			asrRes[i].QueuePosition = position
			asrRes[i].QueueType = queueType
		}
	}

	// 即使查询结果为空也返回空列表
	c.JSON(http.StatusOK, gin.H{
		"total": count,
		"page":  page,
		"size":  size,
		"items": asrRes,
	})
}

// DeleteASRTask 删除ASR任务
func DeleteASRTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	taskID := c.Param("id")

	// 查询任务
	var task models.ASRTask
	result := db.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 删除任务
	if err := db.DB.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	// 删除关联的音频文件
	if task.InputFile != "" {
		// 删除文件
		if err := storages.Client.Delete(context.Background(), task.InputFile); err != nil {
			logrus.Errorf("Error deleting input file: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}
