package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
)

// CreateMessageRequest 创建留言请求
type CreateMessageRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Phone   string `json:"phone"`
	Subject string `json:"subject" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CreateMessage 创建留言
func CreateMessage(c *gin.Context) {
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 创建留言
	message := models.Message{
		UserID:  userID.(uint),
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Subject: req.Subject,
		Content: req.Content,
		Status:  "unread",
	}

	// 保存到数据库
	result := db.DB.Create(&message)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建留言失败: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "留言提交成功",
		"data":    message,
	})
}

// GetMessage 获取单个留言
func GetMessage(c *gin.Context) {
	id := c.Param("id")
	messageID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的留言ID",
		})
		return
	}

	var message models.Message
	result := db.DB.First(&message, messageID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "留言不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取留言成功",
		"data":    message,
	})
}

// ListMessages 获取留言列表
func ListMessages(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 获取用户角色
	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "无法获取用户角色",
		})
		return
	}

	var messages []models.Message
	query := db.DB

	// 如果不是管理员，只能查看自己的留言
	if userRole != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// 获取总数
	var total int64
	query.Model(&models.Message{}).Count(&total)

	// 获取数据
	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&messages)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取留言列表失败: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取留言列表成功",
		"data": gin.H{
			"list":  messages,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// ReplyMessage 回复留言
func ReplyMessage(c *gin.Context) {
	// 检查是否是管理员
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "只有管理员可以回复留言",
		})
		return
	}

	id := c.Param("id")
	messageID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的留言ID",
		})
		return
	}

	// 获取管理员ID
	adminID, _ := c.Get("user_id")

	// 获取回复内容
	var req struct {
		ReplyText string `json:"reply_text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新留言
	result := db.DB.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"status":     "replied",
			"reply_text": req.ReplyText,
			"reply_time": time.Now(),
			"admin_id":   adminID,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "回复留言失败: " + result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "留言不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "回复留言成功",
	})
}

// DeleteMessage 删除留言
func DeleteMessage(c *gin.Context) {
	id := c.Param("id")
	messageID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的留言ID",
		})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 获取用户角色
	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "无法获取用户角色",
		})
		return
	}

	// 构建查询条件
	query := db.DB
	// 如果不是管理员，只能删除自己的留言
	if userRole != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	// 删除留言
	result := query.Delete(&models.Message{}, messageID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除留言失败: " + result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "留言不存在或无权删除",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除留言成功",
	})
}

// MarkMessageAsRead 标记留言为已读
func MarkMessageAsRead(c *gin.Context) {
	// 检查是否是管理员
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "只有管理员可以标记留言为已读",
		})
		return
	}

	id := c.Param("id")
	messageID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的留言ID",
		})
		return
	}

	// 更新留言状态
	result := db.DB.Model(&models.Message{}).
		Where("id = ? AND status = ?", messageID, "unread").
		Update("status", "read")

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "标记留言为已读失败: " + result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "留言不存在或已经被标记为已读/已回复",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "标记留言为已读成功",
	})
}
