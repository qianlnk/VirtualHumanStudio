package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetImageProcessingTask 获取特定图像处理任务的详细信息
func GetImageProcessingTask(c *gin.Context) {
	// 获取模块ID和任务ID
	moduleId := c.Param("moduleId")
	taskId := c.Param("taskId")
	if moduleId == "" || taskId == "" {
		c.JSON(400, gin.H{"error": "模块ID和任务ID不能为空"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未授权"})
		return
	}

	// 查询任务
	var task models.ComfyUIWorkflowTask
	result := db.DB.Where("id = ? AND user_id = ? AND task_type = ?", taskId, userID, moduleId).First(&task)
	if result.Error != nil {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}

	var inputParams []*models.InputParam
	if err := json.Unmarshal([]byte(task.InputParams), &inputParams); err == nil {
		for _, param := range inputParams {
			if param.Type == models.ParamTypeImage || param.Type == models.ParamTypeMask {
				param.Value = utils.GetFileURL(param.Value)
			}
		}
	}

	var outputParams []*models.InputParam
	if err := json.Unmarshal([]byte(task.OutputParams), &outputParams); err == nil {
		for _, param := range outputParams {
			if param.Type == models.ParamTypeImage || param.Type == models.ParamTypeMask || param.Type == models.ParamTypeVideo {
				param.Value = utils.GetFileURL(param.Value)
			}
		}
	}

	task.InputParams = utils.ToJSONString(inputParams)
	task.OutputParams = utils.ToJSONString(outputParams)

	// 返回任务详情
	c.JSON(200, gin.H{
		"success": true,
		"task":    task,
	})
}

// DeleteImageProcessingTask 删除指定的图像处理任务
func DeleteImageProcessingTask(c *gin.Context) {
	// 获取模块ID和任务ID
	moduleId := c.Param("moduleId")
	taskId := c.Param("taskId")
	if moduleId == "" || taskId == "" {
		c.JSON(400, gin.H{"error": "模块ID和任务ID不能为空"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未授权"})
		return
	}

	// 查询任务
	var task models.ComfyUIWorkflowTask
	result := db.DB.Where("id = ? AND user_id = ? AND task_type = ?", taskId, userID, moduleId).First(&task)
	if result.Error != nil {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}

	// 删除任务相关的文件
	var inputParams []*models.InputParam
	if err := json.Unmarshal([]byte(task.InputParams), &inputParams); err == nil {
		for _, param := range inputParams {
			if param.Type == models.ParamTypeImage || param.Type == models.ParamTypeMask {
				filePath := filepath.Join(config.AppConfig.DataDir, param.Value)
				os.Remove(filePath)
			}
		}
	}

	// 删除任务记录
	if err := db.DB.Delete(&task).Error; err != nil {
		c.JSON(500, gin.H{"error": "删除任务失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "任务已删除",
	})
}

// GetImageProcessingTasks 获取图像处理任务列表
func GetImageProcessingTasks(c *gin.Context) {
	// 获取模块ID
	moduleId := c.Param("moduleId")
	if moduleId == "" {
		c.JSON(400, gin.H{"error": "模块ID不能为空"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未授权"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	var count int64
	db.DB.Model(&models.ComfyUIWorkflowTask{}).Where("user_id = ? AND task_type = ?", userID, moduleId).Count(&count)

	// 查询任务列表
	var tasks []models.ComfyUIWorkflowTask
	result := db.DB.Where("user_id = ? AND task_type = ?", userID, moduleId).Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&tasks)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "获取任务列表失败"})
		return
	}

	for i, task := range tasks {
		var inputParams []*models.InputParam
		if err := json.Unmarshal([]byte(task.InputParams), &inputParams); err == nil {
			for _, param := range inputParams {
				if param.Type == models.ParamTypeImage || param.Type == models.ParamTypeMask {
					param.Value = utils.GetFileURL(param.Value)
				}
			}
		}
		var outputParams []*models.InputParam
		if err := json.Unmarshal([]byte(task.OutputParams), &outputParams); err == nil {
			for _, param := range outputParams {
				if param.Type == models.ParamTypeImage || param.Type == models.ParamTypeMask || param.Type == models.ParamTypeVideo {
					param.Value = utils.GetFileURL(param.Value)
				}
			}
		}
		tasks[i].InputParams = utils.ToJSONString(inputParams)
		tasks[i].OutputParams = utils.ToJSONString(outputParams)
	}

	// 返回任务列表
	c.JSON(200, gin.H{
		"success": true,
		"total":   count,
		"page":    page,
		"size":    size,
		"tasks":   tasks,
	})
}

// ImageProcessingTaskRequest 图像处理任务请求
type ImageProcessingTaskRequest struct {
	Name         string `form:"name" binding:"required"`
	Description  string `form:"description"`
	WorkflowName string `form:"workflow_name" binding:"required"` // 工作流名称
	InputParams  string `form:"input_params"`                     // 输入参数列表（JSON字符串）
}

// CreateImageProcessingTask 创建图像处理任务
func CreateImageProcessingTask(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取模块ID
	moduleId := c.Param("moduleId")
	if moduleId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模块ID不能为空"})
		return
	}

	// 解析表单数据
	var req ImageProcessingTaskRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 处理输入参数
	filePaths := make(map[string]string)
	uploadedFiles := make([]string, 0)

	fmt.Println("req: ", utils.ToJSONString(req))

	// 解析输入参数JSON字符串
	var inputParams []*models.InputParam
	if req.InputParams != "" {
		if err := json.Unmarshal([]byte(req.InputParams), &inputParams); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "输入参数解析失败: " + err.Error()})
			return
		}
	}

	// 处理每个输入参数
	for i, param := range inputParams {
		if param.Type == models.ParamTypeImage || param.Type == models.ParamTypeMask {
			// 获取上传的文件
			file, err := c.FormFile(param.Key)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("未提供%s文件", param.Alias)})
				// 清理已上传的文件
				for _, path := range uploadedFiles {
					os.Remove(path)
				}
				return
			}

			// 检查文件类型
			ext := filepath.Ext(file.Filename)
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持PNG或JPG格式的图片文件"})
				// 清理已上传的文件
				for _, path := range uploadedFiles {
					os.Remove(path)
				}
				return
			}

			// 生成文件名并保存
			uniqueID := uuid.New().String()
			fileName := fmt.Sprintf("%s%s", uniqueID, ext)
			filePath := utils.GetUserFilePath(userID.(uint), config.AppConfig.UploadDir, fileName)
			fullPath := filepath.Join(config.AppConfig.DataDir, filePath)

			if err := c.SaveUploadedFile(file, fullPath); err != nil {
				// 清理已上传的文件
				for _, path := range uploadedFiles {
					os.Remove(path)
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存%s文件失败: %s", param.Alias, err.Error())})
				return
			}

			// 记录文件路径
			inputParams[i].Value = filePath
			filePaths[param.Key] = filePath
			uploadedFiles = append(uploadedFiles, fullPath)
		}
	}

	fmt.Println("filePaths: ", filePaths)

	// 创建工作流任务记录
	task := models.ComfyUIWorkflowTask{
		ComfyUITask: models.ComfyUITask{
			UserID:      userID.(uint),
			Name:        req.Name,
			Description: req.Description,
			TaskType:    moduleId,
			Status:      "pending",
		},
		WorkflowName: req.WorkflowName,
		InputParams:  utils.ToJSONString(inputParams),
	}

	// 保存到数据库
	result := db.DB.Create(&task)
	if result.Error != nil {
		// 删除已上传的图片
		for _, path := range uploadedFiles {
			os.Remove(path)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败: " + result.Error.Error()})
		return
	}

	// 异步处理图像处理任务
	go processImageTask(task.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"task":    task,
	})
}

func RetryImageProcessingTask(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	// 获取模块ID
	moduleId := c.Param("moduleId")
	if moduleId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模块ID不能为空"})
		return
	}
	// 获取任务ID
	taskId := c.Param("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}
	// 查询任务
	var task models.ComfyUIWorkflowTask
	result := db.DB.Where("id =? AND user_id =? AND task_type =?", taskId, userID, moduleId).First(&task)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 更新任务状态为pending
	task.Status = "pending"
	db.DB.Save(&task)
	// 异步处理图像处理任务
	go processImageTask(task.ID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已重新提交",
	})
}
