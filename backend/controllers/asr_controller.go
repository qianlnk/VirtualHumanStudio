package controllers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"VirtualHumanStudio/backend/config"
	"VirtualHumanStudio/backend/db"
	"VirtualHumanStudio/backend/models"
	"VirtualHumanStudio/backend/utils"

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
	Text       string           `json:"text"`
	Sentences  []ASRSentence    `json:"sentences"`
	Code       int              `json:"code"`
	TimeCostMs map[string]int64 `json:"time_cost"`
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
	fullFilePath := filepath.Join(config.AppConfig.DataDir, filePath)

	// 创建目标文件
	dst, err := os.Create(fullFilePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 保存文件
	if _, err = io.Copy(dst, reader); err != nil {
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
	now := time.Now()

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

	// 重新打开保存的文件用于ASR请求
	audioFile, err := os.Open(filepath.Join(config.AppConfig.DataDir, filePath))
	if err != nil {
		logger.Errorf("Error opening saved audio file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开音频文件失败: " + err.Error()})
		return
	}
	defer audioFile.Close()

	// 调用ASR服务
	rsp, err := callASRService(req.Model, audioFile, audioFileName)
	if err != nil {
		logger.Errorf("Error calling ASR service: %v", err)
		asrTask.Status = "failed"
		asrTask.ErrorMsg = err.Error()
		db.DB.Save(&asrTask)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 处理ASR响应
	if rsp.Code != 0 {
		if rsp.Code == 1 && strings.Contains(rsp.Text, "音频中没有有效语音") {
			logger.Info("No valid speech found in audio")
			asrTask.Status = "completed"
			asrTask.OutputText = "音频中没有有效语音"
		} else {
			logger.Errorf("ASR service returned error code %d: %s", rsp.Code, rsp.Text)
			asrTask.Status = "failed"
			asrTask.ErrorMsg = rsp.Text
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ASR服务错误: " + rsp.Text})
			return
		}
	} else {
		asrTask.Status = "completed"
		asrTask.OutputText = rsp.Text
	}

	// 更新任务状态
	db.DB.Save(&asrTask)

	// 设置处理时间
	rsp.TimeCostMs["total"] = time.Since(now).Milliseconds()

	c.JSON(http.StatusOK, rsp)
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

	task.InputFile = utils.GetFileURL(task.InputFile)

	c.JSON(http.StatusOK, gin.H{
		"task": task,
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

	for _, task := range tasks {
		task.InputFile = utils.GetFileURL(task.InputFile)
	}

	// 即使查询结果为空也返回空列表
	c.JSON(http.StatusOK, gin.H{
		"total": count,
		"page":  page,
		"size":  size,
		"items": tasks,
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
		filePath := filepath.Join(config.AppConfig.DataDir, task.InputFile)
		if err := os.Remove(filePath); err != nil {
			// 仅记录错误，不影响任务删除
			logrus.Errorf("Error deleting input file: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}
