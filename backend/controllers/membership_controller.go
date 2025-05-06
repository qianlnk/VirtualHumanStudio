package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/middleware"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"gorm.io/gorm"
)

// MembershipController 会员控制器
type MembershipController struct {
	DB *gorm.DB
}

// NewMembershipController 创建会员控制器
func NewMembershipController(db *gorm.DB) *MembershipController {
	// 初始化会员计划
	initMembershipPlans(db)
	return &MembershipController{DB: db}
}

// 初始化会员计划
func initMembershipPlans(db *gorm.DB) {
	var count int64
	db.Model(&models.MembershipPlan{}).Count(&count)
	if count == 0 {
		plans := []models.MembershipPlan{
			{
				Name:         "免费用户",
				Level:        models.MembershipFree,
				Price:        0,
				Duration:     -1, // 永久
				TaskPriority: 0,
				DailyLimit:   5, // 免费用户每日限制5次
				Description:  "免费体验版，功能受限",
				Features:     `["基础功能使用","每日限制5次使用","无优先级","有广告"]`,
				IsActive:     true,
			},
			{
				Name:         "月度会员",
				Level:        models.MembershipMonthly,
				Price:        39.9,
				Duration:     30,
				TaskPriority: 1,
				DailyLimit:   30,
				Description:  "月度会员，限时优惠",
				Features:     `["所有基础功能","每日限制30次使用","普通优先级","无广告"]`,
				IsActive:     true,
			},
			{
				Name:         "季度会员",
				Level:        models.MembershipQuarter,
				Price:        99.9,
				Duration:     90,
				TaskPriority: 2,
				DailyLimit:   100,
				Description:  "季度会员，超值优惠",
				Features:     `["所有高级功能","每日限制100次使用","较高优先级","无广告","专属客服"]`,
				IsActive:     true,
			},
			{
				Name:         "年度会员",
				Level:        models.MembershipYearly,
				Price:        299.9,
				Duration:     365,
				TaskPriority: 3,
				DailyLimit:   -1, // 无限制
				Description:  "年度会员，尊享特权",
				Features:     `["所有高级功能","无使用次数限制","最高优先级","无广告","专属客服","优先体验新功能"]`,
				IsActive:     true,
			},
		}

		for _, plan := range plans {
			db.Create(&plan)
		}
	}
}

// RegisterRoutes 注册路由
func (c *MembershipController) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/membership")
	{
		api.GET("/plans", c.GetMembershipPlans)
		api.GET("/user", middleware.JWTAuth(), c.GetUserMembership)
		api.POST("/purchase", middleware.JWTAuth(), c.PurchaseMembership)
		api.POST("/cancel", middleware.JWTAuth(), c.CancelAutoRenew)
		api.GET("/daily-usage", middleware.JWTAuth(), c.GetDailyUsage)
	}
}

// GetMembershipPlans 获取会员计划列表
func (c *MembershipController) GetMembershipPlans(ctx *gin.Context) {
	var plans []models.MembershipPlan
	if err := c.DB.Where("is_active = ?", true).Find(&plans).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员计划失败"})
		return
	}
	ctx.JSON(http.StatusOK, plans)
}

// GetUserMembership 获取用户会员信息
func (c *MembershipController) GetUserMembership(ctx *gin.Context) {
	userIDInterface, _ := ctx.Get("user_id")
	userID := userIDInterface.(uint)

	var membership models.Membership
	if err := c.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		// 如果用户没有会员记录，创建一个免费会员
		if err == gorm.ErrRecordNotFound {
			membership = models.Membership{
				UserID:       userID,
				Level:        models.MembershipFree,
				StartDate:    time.Now(),
				ExpireDate:   time.Date(9999, 12, 31, 23, 59, 59, 0, time.Local), // 永不过期
				Status:       "active",
				TaskPriority: 0,
				DailyLimit:   5,
			}
			c.DB.Create(&membership)
			ctx.JSON(http.StatusOK, membership)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员信息失败"})
		return
	}

	// 检查会员是否过期
	if membership.Level != models.MembershipFree && time.Now().After(membership.ExpireDate) {
		membership.Level = models.MembershipFree
		membership.Status = "expired"
		membership.TaskPriority = 0
		membership.DailyLimit = 5
		c.DB.Save(&membership)
	}

	ctx.JSON(http.StatusOK, membership)
}

// PurchaseMembership 购买会员
func (c *MembershipController) PurchaseMembership(ctx *gin.Context) {
	userIDInterface, _ := ctx.Get("user_id")
	userID := userIDInterface.(uint)

	var req struct {
		PlanID    uint   `json:"plan_id" binding:"required"`
		PaymentID string `json:"payment_id"`
		AutoRenew bool   `json:"auto_renew"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查询会员计划
	var plan models.MembershipPlan
	if err := c.DB.First(&plan, req.PlanID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "会员计划不存在"})
		return
	}

	// 这里应该有支付验证逻辑，略过...

	// 更新或创建会员记录
	var membership models.Membership
	result := c.DB.Where("user_id = ?", userID).First(&membership)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员信息失败"})
		return
	}

	startDate := time.Now()
	var expireDate time.Time
	if plan.Duration > 0 {
		expireDate = startDate.AddDate(0, 0, plan.Duration)
	} else {
		expireDate = time.Date(9999, 12, 31, 23, 59, 59, 0, time.Local) // 永不过期
	}

	// 如果存在会员记录且未过期，则延长有效期
	if result.Error == nil && membership.Level != models.MembershipFree && time.Now().Before(membership.ExpireDate) {
		expireDate = membership.ExpireDate.AddDate(0, 0, plan.Duration)
	}

	membership.UserID = userID
	membership.Level = plan.Level
	membership.StartDate = startDate
	membership.ExpireDate = expireDate
	membership.AutoRenew = req.AutoRenew
	membership.PaymentID = req.PaymentID
	membership.Status = "active"
	membership.TaskPriority = plan.TaskPriority
	membership.DailyLimit = plan.DailyLimit

	if result.Error == gorm.ErrRecordNotFound {
		if err := c.DB.Create(&membership).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建会员记录失败"})
			return
		}
	} else {
		if err := c.DB.Save(&membership).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新会员记录失败"})
			return
		}
	}

	ctx.JSON(http.StatusOK, membership)
}

// CancelAutoRenew 取消自动续费
func (c *MembershipController) CancelAutoRenew(ctx *gin.Context) {
	userIDInterface, _ := ctx.Get("user_id")
	userID := userIDInterface.(uint)

	var membership models.Membership
	if err := c.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "会员信息不存在"})
		return
	}

	membership.AutoRenew = false
	if err := c.DB.Save(&membership).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "取消自动续费失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "已取消自动续费"})
}

// 创建用户使用量模型
type UserUsage struct {
	UserID     uint      `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	Date       time.Time `json:"date" gorm:"primaryKey;type:date"`
	UsageCount int       `json:"usage_count" gorm:"default:0"`
}

// TableName 设置表名
func (UserUsage) TableName() string {
	return "user_usages"
}

// GetDailyUsage 获取用户今日使用量
func (c *MembershipController) GetDailyUsage(ctx *gin.Context) {
	userIDInterface, _ := ctx.Get("user_id")
	userID := userIDInterface.(uint)

	today := time.Now().Format("2006-01-02")

	var usage UserUsage
	if err := c.DB.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			usage = UserUsage{
				UserID:     userID,
				Date:       time.Now(),
				UsageCount: 0,
			}
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取使用量失败"})
			return
		}
	}

	// 获取用户会员信息
	var membership models.Membership
	if err := c.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员信息失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"usage_count": usage.UsageCount,
		"daily_limit": membership.DailyLimit,
	})
}

// IncrementUsage 增加用户使用量
func (c *MembershipController) IncrementUsage(userID uint) error {
	today := time.Now().Format("2006-01-02")

	// 使用事务保证原子性
	tx := c.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var usage UserUsage
	if err := tx.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			usage = UserUsage{
				UserID:     userID,
				Date:       time.Now(),
				UsageCount: 1,
			}
			if err := tx.Create(&usage).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
			return err
		}
	} else {
		usage.UsageCount++
		if err := tx.Save(&usage).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// CheckUsageLimit 检查用户是否超出使用限制
func (c *MembershipController) CheckUsageLimit(userID uint) (bool, error) {
	var membership models.Membership
	if err := c.DB.Where("user_id = ?", userID).First(&membership).Error; err != nil {
		return false, err
	}

	// 无限制
	if membership.DailyLimit < 0 {
		return true, nil
	}

	today := time.Now().Format("2006-01-02")
	var usage UserUsage
	if err := c.DB.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有今日记录，肯定没超限
			return true, nil
		}
		return false, err
	}

	// 检查是否超限
	return usage.UsageCount < membership.DailyLimit, nil
}
