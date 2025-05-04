package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/middleware"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Register 用户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	result := db.DB.Where("username = ?", req.Username).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询错误"})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		Role:     "user", // 默认角色为普通用户
		Status:   1,      // 默认状态为正常
	}

	result = db.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败: " + result.Error.Error()})
		return
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	// 返回用户信息和令牌
	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"token":   token,
		"user": UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 查询用户
	var user models.User
	result := db.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询错误"})
		}
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "账户已被禁用"})
		return
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 更新登录信息
	user.LastLoginAt = time.Now()
	user.LastLoginIP = c.ClientIP()
	db.DB.Save(&user)

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	// 返回用户信息和令牌
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Logout 用户注销
func Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 注销令牌
	err := middleware.Logout(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注销失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注销成功"})
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 查询用户
	var user models.User
	result := db.DB.First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, gin.H{
		"user": UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	})
}

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取请求参数
	var req struct {
		Email string `json:"email" binding:"omitempty,email"`
		Phone string `json:"phone" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 查询用户
	var user models.User
	result := db.DB.First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 更新用户信息
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	result = db.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	// 返回更新后的用户信息
	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"user": UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	})
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 获取请求参数
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 查询用户
	var user models.User
	result := db.DB.First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 验证旧密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码错误"})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	user.Password = string(hashedPassword)
	result = db.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	// 注销当前令牌，强制重新登录
	err = middleware.Logout(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注销当前会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功，请重新登录"})
}

// ListUsers 管理员获取用户列表
func ListUsers(c *gin.Context) {
	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 查询用户总数
	var count int64
	db.DB.Model(&models.User{}).Count(&count)

	// 查询用户列表
	var users []models.User
	result := db.DB.Offset((page - 1) * size).Limit(size).Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	// 构建响应
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total": count,
		"page":  page,
		"size":  size,
		"users": userResponses,
	})
}

// UpdateUserStatus 管理员更新用户状态
func UpdateUserStatus(c *gin.Context) {
	// 获取用户ID
	userIDStr := c.Param("id")
	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取请求参数
	var req struct {
		Status int `json:"status" binding:"required,oneof=0 1"` // 0: 禁用, 1: 正常
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 查询用户
	var user models.User
	result := db.DB.First(&user, userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		}
		return
	}

	// 更新用户状态
	user.Status = req.Status
	result = db.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户状态失败"})
		return
	}

	// 如果禁用用户，注销其令牌
	if req.Status == 0 {
		err := middleware.Logout(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注销用户会话失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新用户状态成功",
		"user": UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	})
}
