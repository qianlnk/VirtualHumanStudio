package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/storages"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
)

// DigitalHumanTemplateRequest 创建模版请求
// 视频必填，人脸图片/背景图片/视频可选
// multipart/form-data

type DigitalHumanTemplateRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
}

// CreateDigitalHumanTemplate 创建数字人模版
func CreateDigitalHumanTemplate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req DigitalHumanTemplateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 视频文件必填
	videoFile, err := c.FormFile("video_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供视频文件"})
		return
	}
	videoExt := filepath.Ext(videoFile.Filename)
	if videoExt != ".mp4" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持MP4格式的视频文件"})
		return
	}
	videoUniqueID := uuid.New().String()
	videoFileName := fmt.Sprintf("%s%s", videoUniqueID, videoExt)
	videoFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.VideoDir, videoFileName)
	video, err := videoFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开视频文件失败: " + err.Error()})
		return
	}
	err = storages.Client.SaveFile(context.Background(), videoFilePath, video)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存视频文件失败: " + err.Error()})
		return
	}

	// 人脸图片（可选）
	faceImageFile, err := c.FormFile("face_image_file")
	var faceImageFilePath string
	if err == nil {
		faceImageExt := filepath.Ext(faceImageFile.Filename)
		// if faceImageExt != ".jpg" && faceImageExt != ".jpeg" && faceImageExt != ".png" {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持JPG/PNG格式的人脸图片"})
		// 	return
		// }
		faceImageUniqueID := uuid.New().String()
		faceImageFileName := fmt.Sprintf("%s%s", faceImageUniqueID, faceImageExt)
		faceImageFilePath = utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, faceImageFileName)
		faceImage, err := faceImageFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打开人脸图片失败: " + err.Error()})
			return
		}
		err = storages.Client.SaveFile(context.Background(), faceImageFilePath, faceImage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存人脸图片失败: " + err.Error()})
			return
		}
	}

	// 背景图片/视频（可选）
	backgroundFile, err := c.FormFile("background_file")
	var backgroundFilePath string
	if err == nil {
		backgroundExt := filepath.Ext(backgroundFile.Filename)
		if backgroundExt != ".jpg" && backgroundExt != ".jpeg" && backgroundExt != ".png" && backgroundExt != ".mp4" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持JPG/PNG/MP4格式的背景文件"})
			return
		}
		backgroundUniqueID := uuid.New().String()
		backgroundFileName := fmt.Sprintf("%s%s", backgroundUniqueID, backgroundExt)
		backgroundFilePath = utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, backgroundFileName)
		background, err := backgroundFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "打开背景文件失败: " + err.Error()})
			return
		}
		err = storages.Client.SaveFile(context.Background(), backgroundFilePath, background)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存背景文件失败: " + err.Error()})
			return
		}
	}

	template := models.DigitalHumanTemplate{
		UserID:        userID.(uint),
		Name:          req.Name,
		Description:   req.Description,
		VideoURL:      videoFilePath,
		FaceImageURL:  faceImageFilePath,
		BackgroundURL: backgroundFilePath,
		Status:        "pending",
	}
	result := db.DB.Create(&template)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建模版失败: " + result.Error.Error()})
		return
	}

	go processDigitalHumanTemplate(context.Background(), template.ID)

	template.VideoURL, _ = storages.Client.GetFileUrl(context.Background(), template.VideoURL)
	if template.FaceImageURL != "" {
		template.FaceImageURL, _ = storages.Client.GetFileUrl(context.Background(), template.FaceImageURL)
	}
	if template.BackgroundURL != "" {
		template.BackgroundURL, _ = storages.Client.GetFileUrl(context.Background(), template.BackgroundURL)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "数字人模版已创建",
		"template": template,
	})
}

// ListDigitalHumanTemplates 获取模版列表
func ListDigitalHumanTemplates(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	page, size := utils.GetPaginationParams(c)
	var count int64
	db.DB.Model(&models.DigitalHumanTemplate{}).Where("user_id = ?", userID).Count(&count)
	var templates []models.DigitalHumanTemplate
	result := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&templates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}
	for i := range templates {
		templates[i].VideoURL, _ = storages.Client.GetFileUrl(context.Background(), templates[i].VideoURL)
		if templates[i].FaceImageURL != "" {
			templates[i].FaceImageURL, _ = storages.Client.GetFileUrl(context.Background(), templates[i].FaceImageURL)
		}
		if templates[i].BackgroundURL != "" {
			templates[i].BackgroundURL, _ = storages.Client.GetFileUrl(context.Background(), templates[i].BackgroundURL)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"total":     count,
		"page":      page,
		"size":      size,
		"templates": templates,
	})
}

// GetDigitalHumanTemplate 获取模版详情
func GetDigitalHumanTemplate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	id := c.Param("id")
	var template models.DigitalHumanTemplate
	result := db.DB.First(&template, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模版不存在"})
		return
	}
	if template.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此模版"})
		return
	}
	template.VideoURL, _ = storages.Client.GetFileUrl(context.Background(), template.VideoURL)
	if template.FaceImageURL != "" {
		template.FaceImageURL, _ = storages.Client.GetFileUrl(context.Background(), template.FaceImageURL)
	}
	if template.BackgroundURL != "" {
		template.BackgroundURL, _ = storages.Client.GetFileUrl(context.Background(), template.BackgroundURL)
	}
	if template.ReplaceFaceVideoURL != "" {
		template.ReplaceFaceVideoURL, _ = storages.Client.GetFileUrl(context.Background(), template.ReplaceFaceVideoURL)
	}
	if template.ReplaceBackgroundVideoURL != "" {
		template.ReplaceBackgroundVideoURL, _ = storages.Client.GetFileUrl(context.Background(), template.ReplaceBackgroundVideoURL)
	}
	if template.ResultVideoURL != "" {
		template.ResultVideoURL, _ = storages.Client.GetFileUrl(context.Background(), template.ResultVideoURL)
	}
	c.JSON(http.StatusOK, gin.H{"template": template})
}

// DeleteDigitalHumanTemplate 删除模版
func DeleteDigitalHumanTemplate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	id := c.Param("id")
	var template models.DigitalHumanTemplate
	result := db.DB.First(&template, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模版不存在"})
		return
	}
	if template.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除此模版"})
		return
	}
	if template.VideoURL != "" {
		storages.Client.Delete(context.Background(), template.VideoURL)
	}
	if template.FaceImageURL != "" {
		storages.Client.Delete(context.Background(), template.FaceImageURL)
	}
	if template.BackgroundURL != "" {
		storages.Client.Delete(context.Background(), template.BackgroundURL)
	}
	result = db.DB.Delete(&template)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "数字人模版已删除"})
}

func processDigitalHumanTemplate(ctx context.Context, tempID uint) error {
	// 获取数字人模版
	var temp models.DigitalHumanTemplate
	if err := db.DB.First(&temp, tempID).Error; err != nil {
		fmt.Printf("获取数字人模版失败: %v\n", err)
		return err
	}

	// 更新数字人模版状态
	temp.Status = "processing"
	temp.ResultVideoURL = temp.VideoURL
	db.DB.Save(&temp)

	if temp.FaceImageURL != "" && temp.BackgroundURL != "" { // 换脸换背景
		// 先换脸
		video, err := replaceVideoFace(ctx, &temp)
		if err != nil {
			temp.Status = "failed"
			temp.ErrorMsg = err.Error()
			db.DB.Save(&temp)
			return err
		}
		temp.ReplaceFaceVideoURL = video
		temp.ResultVideoURL = video
		temp.Status = "completed"

		// 再换背景
		video, err = replaceVideoBackgrpund(ctx, &temp)
		if err != nil {
			temp.Status = "failed"
			temp.ErrorMsg = err.Error()
			db.DB.Save(&temp)
			return err
		}
		temp.ReplaceBackgroundVideoURL = video
		temp.ResultVideoURL = video
		temp.Status = "completed"
		db.DB.Save(&temp)

	} else if temp.FaceImageURL != "" { // 只换脸
		// 换脸
		video, err := replaceVideoFace(ctx, &temp)
		if err != nil {
			temp.Status = "failed"
			temp.ErrorMsg = err.Error()
			db.DB.Save(&temp)
			return err
		}
		temp.ReplaceFaceVideoURL = video
		temp.ResultVideoURL = video
		temp.Status = "completed"
		db.DB.Save(&temp)
	} else if temp.BackgroundURL != "" { // 只换背景
		// 更换背景
		video, err := replaceVideoBackgrpund(ctx, &temp)
		if err != nil {
			temp.Status = "failed"
			temp.ErrorMsg = err.Error()
			db.DB.Save(&temp)
			return err
		}
		temp.ReplaceBackgroundVideoURL = video
		temp.ResultVideoURL = video
		temp.Status = "completed"
		db.DB.Save(&temp)
	}

	// 啥也不换
	temp.Status = "completed"
	db.DB.Save(&temp)

	return nil
}

func replaceVideoFace(ctx context.Context, temp *models.DigitalHumanTemplate) (string, error) {
	// 构建换脸任务
	task := &models.ComfyUIWorkflowTask{}

	task.TaskType = "video_replace_face"
	task.WorkflowName = "video_replace_face"
	task.Status = "pending"
	task.UserID = temp.UserID
	task.InputParams = utils.ToJSONString([]models.InputParam{
		{
			Key:   "video",
			Type:  models.ParamTypeVideo,
			Value: temp.VideoURL,
			Alias: "视频",
		},
		{
			Key:   "peopleImage",
			Type:  models.ParamTypeImage,
			Value: temp.FaceImageURL,
			Alias: "人脸",
		},
	})

	// 发起换脸
	task, err := processImageTaskWithInfo(ctx, task)
	if err != nil {
		return "", err
	}

	task, err = waitProcessImageTaskWithInfoCompleted(ctx, task)
	if err != nil {
		return "", err
	}

	var outputParams []models.InputParam
	err = json.Unmarshal([]byte(task.OutputParams), &outputParams)
	if err != nil {
		return "", err
	}

	return outputParams[0].Value, nil
}

func videoMattingToGreen(ctx context.Context, temp *models.DigitalHumanTemplate) (string, error) {
	// 构建视频抠图转绿幕任务
	task := &models.ComfyUIWorkflowTask{}

	video := temp.VideoURL
	// 如果换脸了，则使用换脸后的视频，再换背景
	if temp.ReplaceFaceVideoURL != "" {
		video = temp.ReplaceFaceVideoURL
	}

	task.TaskType = "video_matting_to_green"
	task.WorkflowName = "video_matting_to_green"
	task.Status = "pending"
	task.UserID = temp.UserID
	task.InputParams = utils.ToJSONString([]models.InputParam{
		{
			Key:   "video",
			Type:  models.ParamTypeVideo,
			Value: video,
			Alias: "视频",
		},
		{
			Key:   "backgroundColor",
			Type:  models.ParamTypeText,
			Value: "#00A03C",
			Alias: "背景颜色",
		},
	})

	// 发起任务
	task, err := processImageTaskWithInfo(ctx, task)
	if err != nil {
		return "", err
	}

	task, err = waitProcessImageTaskWithInfoCompleted(ctx, task)
	if err != nil {
		return "", err
	}

	var outputParams []models.InputParam
	err = json.Unmarshal([]byte(task.OutputParams), &outputParams)
	if err != nil {
		return "", err
	}

	return outputParams[0].Value, nil
}

func replaceVideoBackgrpund(ctx context.Context, temp *models.DigitalHumanTemplate) (string, error) {
	// 先抠图转绿幕
	video, err := videoMattingToGreen(ctx, temp)
	if err != nil {
		return "", err
	}

	background := temp.BackgroundURL

	ext := filepath.Ext(video)
	inputPath := "/tmp/" + uuid.New().String() + ext
	tmpPath := "/tmp/" + uuid.New().String() + ext
	obsFileName := uuid.New().String() + ext
	obsFileName = filepath.Join(filepath.Dir(video), obsFileName)

	ext = filepath.Ext(background)
	inputBackgroundPath := "/tmp/" + uuid.New().String() + ext

	videoFile, err := storages.Client.OpenFile(context.Background(), video)
	if err != nil {
		return "", err
	}

	inputFile, err := os.Create(inputPath)
	if err != nil {
		return "", err
	}

	io.Copy(inputFile, videoFile)
	videoFile.Close()
	inputFile.Close()

	backgroundFile, err := storages.Client.OpenFile(context.Background(), background)
	if err != nil {
		return "", err
	}
	inputBackgroundFile, err := os.Create(inputBackgroundPath)
	if err != nil {
		return "", err
	}

	io.Copy(inputBackgroundFile, backgroundFile)
	backgroundFile.Close()
	inputBackgroundFile.Close()

	// 首先获取视频尺寸信息
	probeCmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		inputPath,
	)

	output, err := probeCmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取视频尺寸失败: %v", err)
	}

	// 解析视频尺寸 (格式: width,height)
	var width, height int
	_, err = fmt.Sscanf(string(output), "%d,%d", &width, &height)
	if err != nil {
		return "", fmt.Errorf("解析视频尺寸失败: %v", err)
	}

	fmt.Printf("视频尺寸: %dx%d\n", width, height)

	// 优化背景替换方法：
	// 1. 背景图片使用 cover 模式（填满整个区域）
	// 2. 使用多种绿色值进行抠像，处理光线不均匀
	// 3. 使用更宽松的参数处理强光区域
	greenColor := "0x00A03C" // 绿色

	fmt.Printf("尝试绿色值 %s...\n", greenColor)

	filterComplex := fmt.Sprintf(
		"[0:v]setpts=PTS-STARTPTS[main];"+
			"[1:v]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,"+
			"trim=start=0:duration=0.1,loop=loop=-1:size=1[bg];"+
			"[main]colorkey=%s:0.15:0.05,format=yuva420p[ckout];"+
			"[bg][ckout]overlay=eof_action=pass:shortest=1",
		width, height, width, height, greenColor,
	)

	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-loop", "1",
		"-framerate", "60",
		"-i", inputBackgroundPath,
		"-filter_complex", filterComplex,
		"-c:a", "copy",
		"-preset", "fast",
		"-crf", "20",
		"-movflags", "+faststart",
		tmpPath,
	)

	fmt.Println("执行命令:", cmd.String())
	if err := cmd.Run(); err != nil {
		fmt.Printf("绿色值 %s 执行失败: %v\n", greenColor, err)
		os.Remove(tmpPath)
		return "", err
	}

	// 转存到obs
	tempFile, err := os.Open(tmpPath)
	if err != nil {
		return "", err
	}
	storages.Client.SaveFile(context.Background(), obsFileName, tempFile)
	tempFile.Close()

	os.Remove(inputPath)
	os.Remove(inputBackgroundPath)
	os.Remove(tmpPath)

	return obsFileName, nil
}
