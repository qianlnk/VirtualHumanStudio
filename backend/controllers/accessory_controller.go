package controllers

import (
	"fmt"
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
)

// AccessoryRequest 物品替换请求
type AccessoryRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
	// 物品图、模特图和蒙版图通过multipart/form-data上传
}

// AccessoryResponse 物品替换响应
type AccessoryResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ItemImage   string    `json:"item_image"`
	ModelImage  string    `json:"model_image"`
	MaskImage   string    `json:"mask_image"`
	ResultImage string    `json:"result_image"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAccessory 创建物品替换任务
func CreateAccessory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 解析表单数据
	var req AccessoryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 获取上传的物品图片
	itemFile, err := c.FormFile("white_image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供物品白底图片"})
		return
	}

	// 检查图片文件类型
	itemExt := filepath.Ext(itemFile.Filename)
	if itemExt != ".png" && itemExt != ".jpg" && itemExt != ".jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持PNG或JPG格式的图片文件"})
		return
	}

	// 获取上传的模特图片
	modelFile, err := c.FormFile("model_image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供模特图片"})
		return
	}

	// 检查图片文件类型
	modelExt := filepath.Ext(modelFile.Filename)
	if modelExt != ".png" && modelExt != ".jpg" && modelExt != ".jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持PNG或JPG格式的图片文件"})
		return
	}

	// 获取上传的蒙版图片
	maskFile, err := c.FormFile("mask_image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供蒙版图片"})
		return
	}

	// 检查图片文件类型
	maskExt := filepath.Ext(maskFile.Filename)
	if maskExt != ".png" && maskExt != ".jpg" && maskExt != ".jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持PNG或JPG格式的图片文件"})
		return
	}

	// 保存物品图片
	itemUniqueID := uuid.New().String()
	itemFileName := fmt.Sprintf("%s%s", itemUniqueID, itemExt)
	itemFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, itemFileName)
	itemFullPath := filepath.Join(config.AppConfig.DataDir, itemFilePath)

	if err := c.SaveUploadedFile(itemFile, itemFullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存物品白底图片失败: " + err.Error()})
		return
	}

	// 保存模特图片
	modelUniqueID := uuid.New().String()
	modelFileName := fmt.Sprintf("%s%s", modelUniqueID, modelExt)
	modelFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, modelFileName)
	modelFullPath := filepath.Join(config.AppConfig.DataDir, modelFilePath)

	if err := c.SaveUploadedFile(modelFile, modelFullPath); err != nil {
		// 删除已上传的物品图片
		os.Remove(itemFullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存模特图片失败: " + err.Error()})
		return
	}

	// 保存蒙版图片
	maskUniqueID := uuid.New().String()
	maskFileName := fmt.Sprintf("%s%s", maskUniqueID, maskExt)
	maskFilePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, maskFileName)
	maskFullPath := filepath.Join(config.AppConfig.DataDir, maskFilePath)

	if err := c.SaveUploadedFile(maskFile, maskFullPath); err != nil {
		// 删除已上传的图片
		os.Remove(itemFullPath)
		os.Remove(modelFullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存蒙版图片失败: " + err.Error()})
		return
	}

	// 创建物品替换任务记录
	accessory := models.Accessory{
		ComfyUITask: models.ComfyUITask{
			UserID:      userID.(uint),
			Name:        req.Name,
			Description: req.Description,
			TaskType:    "accessory",
			Status:      "pending",
		},
		ItemImage:  itemFilePath,
		ModelImage: modelFilePath,
		MaskImage:  maskFilePath,
	}

	result := db.DB.Create(&accessory)
	if result.Error != nil {
		// 删除已上传的图片
		os.Remove(itemFullPath)
		os.Remove(modelFullPath)
		os.Remove(maskFullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建物品替换任务记录失败: " + result.Error.Error()})
		return
	}

	// 异步处理物品替换任务
	go processAccessory(accessory.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "物品替换任务创建成功",
		"accessory": AccessoryResponse{
			ID:          accessory.ID,
			Name:        accessory.Name,
			Description: accessory.Description,
			ItemImage:   utils.GetFileURL(accessory.ItemImage),
			ModelImage:  utils.GetFileURL(accessory.ModelImage),
			MaskImage:   utils.GetFileURL(accessory.MaskImage),
			Status:      accessory.Status,
			CreatedAt:   accessory.CreatedAt,
			UpdatedAt:   accessory.UpdatedAt,
		},
	})
}

// GetAccessory 获取单个物品替换任务详情
func GetAccessory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务记录
	var accessory models.Accessory
	result := db.DB.Where("user_id = ?", userID).First(&accessory, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到物品替换任务"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessory": AccessoryResponse{
			ID:          accessory.ID,
			Name:        accessory.Name,
			Description: accessory.Description,
			ItemImage:   utils.GetFileURL(accessory.ItemImage),
			ModelImage:  utils.GetFileURL(accessory.ModelImage),
			MaskImage:   utils.GetFileURL(accessory.MaskImage),
			ResultImage: utils.GetFileURL(accessory.ResultImage),
			Status:      accessory.Status,
			CreatedAt:   accessory.CreatedAt,
			UpdatedAt:   accessory.UpdatedAt,
		},
	})
}

// ListAccessories 获取物品替换任务列表
func ListAccessories(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 查询条件
	query := db.DB.Where("user_id = ?", userID)

	// 查询总数
	var count int64
	query.Model(&models.Accessory{}).Count(&count)

	// 查询列表
	var accessories []models.Accessory
	result := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&accessories)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 构建响应
	responses := make([]AccessoryResponse, len(accessories))
	for i, a := range accessories {
		responses[i] = AccessoryResponse{
			ID:          a.ID,
			Name:        a.Name,
			Description: a.Description,
			ItemImage:   utils.GetFileURL(a.ItemImage),
			ModelImage:  utils.GetFileURL(a.ModelImage),
			MaskImage:   utils.GetFileURL(a.MaskImage),
			ResultImage: utils.GetFileURL(a.ResultImage),
			Status:      a.Status,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":       count,
		"page":        page,
		"size":        size,
		"accessories": responses,
	})
}

// DeleteAccessory 删除物品替换任务
func DeleteAccessory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取任务ID
	id := c.Param("id")

	// 查询任务记录
	var accessory models.Accessory
	result := db.DB.Where("user_id = ?", userID).First(&accessory, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到物品替换任务"})
		return
	}

	// 删除相关文件
	if accessory.ItemImage != "" {
		os.Remove(filepath.Join(config.AppConfig.DataDir, accessory.ItemImage))
	}
	if accessory.ModelImage != "" {
		os.Remove(filepath.Join(config.AppConfig.DataDir, accessory.ModelImage))
	}
	if accessory.MaskImage != "" {
		os.Remove(filepath.Join(config.AppConfig.DataDir, accessory.MaskImage))
	}
	if accessory.ResultImage != "" {
		os.Remove(filepath.Join(config.AppConfig.DataDir, accessory.ResultImage))
	}

	// 删除数据库记录
	result = db.DB.Delete(&accessory)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// processAccessory 处理物品替换任务
func processAccessory(accessoryID uint) {
	// 查询任务记录
	var accessory models.Accessory
	result := db.DB.First(&accessory, accessoryID)
	if result.Error != nil {
		fmt.Printf("查询物品替换任务失败: %v\n", result.Error)
		return
	}

	// 更新任务状态为处理中
	db.DB.Model(&accessory).Update("status", "processing")

	// 上传物品图片到ComfyUI服务器
	itemImageRef, err := UploadImageToComfyUI(filepath.Join(config.AppConfig.DataDir, accessory.ItemImage))
	if err != nil {
		fmt.Printf("上传物品图片到ComfyUI服务器失败: %v\n", err)
		db.DB.Model(&accessory).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "上传物品图片到ComfyUI服务器失败: " + err.Error(),
		})
		return
	}

	// 更新物品图片引用
	db.DB.Model(&accessory).Update("item_image_ref", itemImageRef)

	// 上传模特图片到ComfyUI服务器
	modelImageRef, err := UploadImageToComfyUI(filepath.Join(config.AppConfig.DataDir, accessory.ModelImage))
	if err != nil {
		fmt.Printf("上传模特图片到ComfyUI服务器失败: %v\n", err)
		db.DB.Model(&accessory).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "上传模特图片到ComfyUI服务器失败: " + err.Error(),
		})
		return
	}

	// 更新模特图片引用
	db.DB.Model(&accessory).Update("model_image_ref", modelImageRef)

	// 上传蒙版图片到ComfyUI服务器
	maskImageRef, err := UploadMaskToComfyUI(filepath.Join(config.AppConfig.DataDir, accessory.MaskImage), modelImageRef)
	if err != nil {
		fmt.Printf("上传蒙版图片到ComfyUI服务器失败: %v\n", err)
		db.DB.Model(&accessory).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "上传蒙版图片到ComfyUI服务器失败: " + err.Error(),
		})
		return
	}

	// 更新蒙版图片引用
	db.DB.Model(&accessory).Update("mask_image_ref", maskImageRef)

	// 构建物品替换工作流
	prompt := BuildAccessoryPrompt(itemImageRef, modelImageRef, maskImageRef)

	// 更新prompt
	db.DB.Model(&accessory).Update("prompt", prompt)

	// 提交任务到ComfyUI服务器
	promptID, err := SubmitPromptToComfyUI(prompt)
	if err != nil {
		fmt.Printf("提交任务到ComfyUI服务器失败: %v\n", err)
		db.DB.Model(&accessory).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": "提交任务到ComfyUI服务器失败: " + err.Error(),
		})
		return
	}

	// 更新任务ID
	db.DB.Model(&accessory).Update("task_id", promptID)

	// 轮询任务进度
	for {
		// 查询任务进度
		status, data, err := QueryComfyUITaskStatus(promptID)
		if err != nil {
			fmt.Printf("查询任务进度失败: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if status == "completed" {
			// 任务完成，处理结果
			// 生成唯一文件名
			uniqueID := uuid.New().String()
			resultFileName := fmt.Sprintf("%s.png", uniqueID)
			resultFilePath := utils.GetUserFilePath(accessory.UserID, config.AppConfig.UploadDir, resultFileName)
			resultFullPath := filepath.Join(config.AppConfig.DataDir, resultFilePath)

			// 下载结果图片
			err := DownloadComfyUIResult(data, resultFullPath, "54")
			if err != nil {
				fmt.Printf("处理任务结果失败: %v\n", err)
				db.DB.Model(&accessory).Updates(map[string]interface{}{
					"status":    "failed",
					"error_msg": "处理任务结果失败: " + err.Error(),
				})
				return
			}

			// 更新任务状态和结果图片URL
			db.DB.Model(&accessory).Updates(map[string]interface{}{
				"status":       "completed",
				"result_image": resultFilePath,
			})
			break
		} else if status == "failed" {
			// 任务失败
			db.DB.Model(&accessory).Updates(map[string]interface{}{
				"status":    "failed",
				"error_msg": "ComfyUI任务处理失败",
			})
			break
		}

		// 等待一段时间后再次查询
		time.Sleep(5 * time.Second)
	}
}
