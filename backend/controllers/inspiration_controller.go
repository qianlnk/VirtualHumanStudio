package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
)

// InspirationParam 参数结构
type InspirationParam struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// InspirationTaskResponse 共享任务响应
type InspirationTaskResponse struct {
	ID           uint               `json:"id"`
	Type         string             `json:"type"`          // 任务类型: "comfyui" 或 "digital_human"
	TaskType     string             `json:"task_type"`     // 具体任务类型，如"accessory"、"workflow"等
	Name         string             `json:"name"`          // 任务名称
	Description  string             `json:"description"`   // 任务描述
	ResultURL    string             `json:"result_url"`    // 结果图片/视频URL
	InputParams  []InspirationParam `json:"input_params"`  // 输入参数
	OutputParams []InspirationParam `json:"output_params"` // 输出参数
	UserID       uint               `json:"user_id"`       // 创建者ID
	Username     string             `json:"username"`      // 创建者用户名
	CreatedAt    string             `json:"created_at"`    // 创建时间
}

// GetInspirationTasks 获取灵感页面任务
func GetInspirationTasks(c *gin.Context) {
	// 分页参数
	page, size := utils.GetPaginationParams(c)

	shareTasks := []*models.ShareTask{}

	var responses []InspirationTaskResponse
	var totalCount int64

	// 查询总数
	db.DB.Model(&models.ShareTask{}).
		Where("share_status = ?", "approved").
		Count(&totalCount)

	db.DB.Where("share_status = ?", "approved").
		Order("id DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&shareTasks)

	for _, task := range shareTasks {
		fmt.Println("task:", utils.ToJSONString(task))
		if task.Mode == "comfyui" {
			var comfyUITask models.ComfyUIWorkflowTask
			db.DB.Where("id = ?", task.TaskID).First(&comfyUITask)

			responses = append(responses, convertComfyUITaskToResponse(comfyUITask))

		} else if task.Mode == "digital_human" {
			var digitalHuman models.DigitalHuman
			db.DB.Where("id = ?", task.TaskID).First(&digitalHuman)
			responses = append(responses, convertDigitalHumanToResponse(digitalHuman))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": responses,
		"total": totalCount,
		"page":  page,
		"size":  size,
	})
}

// GetInspirationDetail 获取单个灵感详情
func GetInspirationDetail(c *gin.Context) {
	id := c.Param("id")
	taskID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	// 尝试查找ComfyUI任务
	var comfyUITask models.ComfyUIWorkflowTask
	if err := db.DB.Where("id = ? AND share_status = ?", taskID, "approved").First(&comfyUITask).Error; err == nil {
		// 找到了ComfyUI任务，转换为响应格式
		response := convertComfyUITaskToResponse(comfyUITask)
		c.JSON(http.StatusOK, response)
		return
	}

	// 尝试查找数字人任务
	var digitalHuman models.DigitalHuman
	if err := db.DB.Where("id = ? AND share_status = ?", taskID, "approved").First(&digitalHuman).Error; err == nil {
		// 找到了数字人任务，转换为响应格式
		response := convertDigitalHumanToResponse(digitalHuman)
		c.JSON(http.StatusOK, response)
		return
	}

	// 未找到任务
	c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定任务"})
}

// 从ComfyUI工作流任务转换为响应结构体
func convertComfyUITaskToResponse(task models.ComfyUIWorkflowTask) InspirationTaskResponse {
	var resultURL string
	var inputParams, outputParams []InspirationParam

	if task.TaskType == "accessory" {
		var accessory models.Accessory
		db.DB.Where("id = ?", task.ID).First(&accessory)
		resultURL = utils.GetFileURL(accessory.ResultImage)

		// 设置输入参数
		inputParams = []InspirationParam{
			{Key: "item_image", Label: "物品图", Type: "image", Value: utils.GetFileURL(accessory.ItemImage)},
			{Key: "model_image", Label: "模特图", Type: "image", Value: utils.GetFileURL(accessory.ModelImage)},
			{Key: "mask_image", Label: "蒙版图", Type: "mask", Value: utils.GetFileURL(accessory.MaskImage)},
		}

		// 设置输出参数
		outputParams = []InspirationParam{
			{Key: "result_image", Label: "结果图", Type: "image", Value: utils.GetFileURL(accessory.ResultImage)},
		}
	} else {
		// 查询工作流任务获取更多信息
		var workflowTask models.ComfyUIWorkflowTask
		if db.DB.Where("id = ?", task.ID).First(&workflowTask).Error == nil {
			// 解析输入参数
			if workflowTask.InputParams != "" {
				var params []models.InputParam
				json.Unmarshal([]byte(workflowTask.InputParams), &params)

				for _, param := range params {
					inputParam := InspirationParam{
						Key:   param.Key,
						Label: param.Alias,
						Type:  param.Type,
						Value: param.Value,
					}

					// 对于图片类型参数，生成完整URL
					if param.Type == "image" || param.Type == "mask" || param.Type == "video" || param.Type == "audio" {
						inputParam.Value = utils.GetFileURL(param.Value)
					}

					inputParams = append(inputParams, inputParam)
				}
			}

			// 解析输出参数获取结果图片
			if workflowTask.OutputParams != "" {
				json.Unmarshal([]byte(workflowTask.OutputParams), &outputParams)

				// 设置结果URL为第一个图片类型的输出参数
				for i, param := range outputParams {
					if param.Type == "image" || param.Type == "mask" || param.Type == "video" || param.Type == "audio" {
						resultURL = utils.GetFileURL(param.Value)
						outputParams[i].Value = utils.GetFileURL(param.Value)
					}
				}
			}
		}
	}

	// 获取用户名
	username := "用户"
	if task.UserID > 0 {
		var user models.User
		db.DB.Select("username").Where("id = ?", task.UserID).First(&user)
		if user.Username != "" {
			username = user.Username
		}
	}

	return InspirationTaskResponse{
		ID:           task.ID,
		Type:         "comfyui",
		TaskType:     task.TaskType,
		Name:         task.Name,
		Description:  task.Description,
		ResultURL:    resultURL,
		InputParams:  inputParams,
		OutputParams: outputParams,
		UserID:       task.UserID,
		Username:     username,
		CreatedAt:    task.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// 从数字人任务转换为响应结构体
func convertDigitalHumanToResponse(dh models.DigitalHuman) InspirationTaskResponse {
	// 获取用户名
	username := "用户"
	if dh.UserID > 0 {
		var user models.User
		db.DB.Select("username").Where("id = ?", dh.UserID).First(&user)
		if user.Username != "" {
			username = user.Username
		}
	}

	// 适配数字人任务为与图像处理任务相同的格式
	inputParams := []InspirationParam{
		{Key: "audio_file", Label: "音频文件", Type: "audio", Value: utils.GetFileURL(dh.AudioURL)},
		{Key: "video_file", Label: "视频文件", Type: "video", Value: utils.GetFileURL(dh.VideoURL)},
	}

	// 添加其他参数
	if dh.Chaofen > 0 {
		inputParams = append(inputParams, InspirationParam{
			Key: "chaofen", Label: "超分", Type: "text", Value: "开启",
		})
	}

	if dh.WatermarkSwitch > 0 {
		inputParams = append(inputParams, InspirationParam{
			Key: "watermark", Label: "水印", Type: "text", Value: "开启",
		})
	}

	inputParams = append(inputParams, InspirationParam{
		Key: "pn", Label: "PN值", Type: "text", Value: utils.ToJSONString(dh.PN),
	})

	// 设置输出参数
	outputParams := []InspirationParam{
		{Key: "result_video", Label: "合成视频", Type: "video", Value: utils.GetFileURL(dh.ResultURL)},
	}

	return InspirationTaskResponse{
		ID:           dh.ID,
		Type:         "digital_human",
		TaskType:     "digital_human",
		Name:         dh.Name,
		Description:  dh.Description,
		ResultURL:    utils.GetFileURL(dh.ResultURL),
		InputParams:  inputParams,
		OutputParams: outputParams,
		UserID:       dh.UserID,
		Username:     username,
		CreatedAt:    dh.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

type ShareTaskID struct {
	ShareTaskID uint `json:"share_task_id"`
}

// CreateShareTaskLike 创建点赞
func CreateShareTaskLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var shareTaskID ShareTaskID
	if err := c.ShouldBindJSON(&shareTaskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	err := db.DB.Create(&models.ShareTaskLike{
		UserID:      userID.(uint),
		ShareTaskID: shareTaskID.ShareTaskID,
	}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "点赞成功",
	})
}

// DeleteShareTaskLike 删除点赞
func DeleteShareTaskLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	shareTaskID := c.Param("share_task_id")
	shareTaskIDInt, err := strconv.Atoi(shareTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	err = db.DB.Where("user_id = ? AND share_task_id = ?", userID.(uint), uint(shareTaskIDInt)).Delete(&models.ShareTaskLike{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "取消点赞成功",
	})
}

// CreateShareTaskFavorite 创建收藏
func CreateShareTaskFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var shareTaskID ShareTaskID
	if err := c.ShouldBindJSON(&shareTaskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	err := db.DB.Create(&models.ShareTaskFavorite{
		UserID:      userID.(uint),
		ShareTaskID: shareTaskID.ShareTaskID,
	}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "收藏成功",
	})
}

// DeleteShareTaskFavorite 删除收藏
func DeleteShareTaskFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	shareTaskID := c.Param("share_task_id")
	shareTaskIDInt, err := strconv.Atoi(shareTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	err = db.DB.Where("user_id = ? AND share_task_id = ?", userID.(uint), uint(shareTaskIDInt)).Delete(&models.ShareTaskFavorite{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "取消收藏成功",
	})
}

// CreateShareTaskComment 创建评论
func CreateShareTaskComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var shareTaskComment = struct {
		ShareTaskID uint   `json:"share_task_id"`
		Content     string `json:"content"`
	}{
		ShareTaskID: 0,
		Content:     "",
	}

	if err := c.ShouldBindJSON(&shareTaskComment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	var user models.User
	db.DB.Where("id = ?", userID.(uint)).First(&user)

	err := db.DB.Create(&models.ShareTaskComment{
		UserID:      userID.(uint),
		ShareTaskID: shareTaskComment.ShareTaskID,
		Content:     shareTaskComment.Content,
	}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评论失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "评论成功",
	})
}

// DeleteShareTaskComment 删除评论
func DeleteShareTaskComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	shareTaskCommentID := c.Param("share_task_comment_id")
	shareTaskCommentIDInt, err := strconv.Atoi(shareTaskCommentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	err = db.DB.Where("id = ? AND user_id = ?", uint(shareTaskCommentIDInt), userID.(uint)).Delete(&models.ShareTaskComment{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除评论失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除评论成功",
	})
}

// GetShareTaskLikes 获取点赞列表
func GetShareTaskLikes(c *gin.Context) {
	shareTaskID := c.Param("share_task_id")
	shareTaskIDInt, err := strconv.Atoi(shareTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	likes := []*models.ShareTaskLike{}
	db.DB.Where("share_task_id = ?", uint(shareTaskIDInt)).Find(&likes)

	users := []*models.User{}

	var userIDs []uint
	for _, like := range likes {
		userIDs = append(userIDs, like.UserID)
	}

	err = db.DB.Where("id IN (?)", userIDs).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	userInfos := []*models.UserInfo{}
	for _, user := range users {
		userInfos = append(userInfos, &models.UserInfo{
			UserID:   user.ID,
			Username: user.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_infos": userInfos,
	})
}

// GetShareTaskFavorites 获取收藏列表
func GetShareTaskFavorites(c *gin.Context) {
	shareTaskID := c.Param("share_task_id")
	shareTaskIDInt, err := strconv.Atoi(shareTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	favorites := []*models.ShareTaskFavorite{}
	db.DB.Where("share_task_id = ?", uint(shareTaskIDInt)).Find(&favorites)

	users := []*models.User{}

	var userIDs []uint
	for _, favorite := range favorites {
		userIDs = append(userIDs, favorite.UserID)
	}

	err = db.DB.Where("id IN (?)", userIDs).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	userInfos := []*models.UserInfo{}
	for _, user := range users {
		userInfos = append(userInfos, &models.UserInfo{
			UserID:   user.ID,
			Username: user.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_infos": userInfos,
	})
}

// GetShareTaskComments 获取评论列表
func GetShareTaskComments(c *gin.Context) {
	shareTaskID := c.Param("share_task_id")
	shareTaskIDInt, err := strconv.Atoi(shareTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	comments := []*models.ShareTaskComment{}
	db.DB.Where("share_task_id = ? AND deleted_at IS NULL", uint(shareTaskIDInt)).Find(&comments)

	users := []*models.User{}

	var userIDs []uint
	for _, comment := range comments {
		userIDs = append(userIDs, comment.UserID)
	}

	err = db.DB.Where("id IN (?)", userIDs).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	userInfos := []*models.UserInfo{}
	for _, user := range users {
		userInfos = append(userInfos, &models.UserInfo{
			UserID:   user.ID,
			Username: user.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"comments":   comments,
		"user_infos": userInfos,
	})

}
