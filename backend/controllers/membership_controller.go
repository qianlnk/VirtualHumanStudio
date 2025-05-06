package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"
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
				WechatQRCode: "pay/wx39.jpeg",
				AlipayQRCode: "pay/zfb39.jpeg",
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
				WechatQRCode: "pay/wx99.jpeg",
				AlipayQRCode: "pay/zfb99.jpeg",
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
				WechatQRCode: "pay/wx299.jpeg",
				AlipayQRCode: "pay/zfb299.jpeg",
			},
		}

		for _, plan := range plans {
			db.Create(&plan)
		}
	}
}

// GetMembershipPlans 获取会员计划列表
func (c *MembershipController) GetMembershipPlans(ctx *gin.Context) {
	var plans []models.MembershipPlan
	if err := c.DB.Where("is_active = ?", true).Find(&plans).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员计划失败"})
		return
	}

	for i, plan := range plans {
		plans[i].WechatQRCode = utils.GetFileURL(plan.WechatQRCode)
		plans[i].AlipayQRCode = utils.GetFileURL(plan.AlipayQRCode)
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
		PlanID        uint   `json:"plan_id" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required"`
		PaymentRemark string `json:"payment_remark"`
		AutoRenew     bool   `json:"auto_renew"`
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

	// 创建会员订单
	order := models.MembershipOrder{
		UserID:        userID,
		PlanID:        plan.ID,
		PlanName:      plan.Name,
		Level:         plan.Level,
		Price:         plan.Price,
		Duration:      plan.Duration,
		Status:        models.OrderStatusPending,
		PaymentMethod: models.PaymentMethod(req.PaymentMethod),
		PaymentRemark: req.PaymentRemark,
		AutoRenew:     req.AutoRenew,
	}

	if err := c.DB.Create(&order).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// GetUserPendingOrders 获取用户待审核订单
func (c *MembershipController) GetUserPendingOrders(ctx *gin.Context) {
	userIDInterface, _ := ctx.Get("user_id")
	userID := userIDInterface.(uint)

	var orders []models.MembershipOrder
	if err := c.DB.Where("user_id = ? AND status = ?", userID, models.OrderStatusPending).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取待审核订单失败"})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// GetAllPendingOrders 管理员获取所有待审核订单
func (c *MembershipController) GetAllPendingOrders(ctx *gin.Context) {
	// 验证是否是管理员
	roleInterface, exists := ctx.Get("role")
	if !exists || roleInterface.(string) != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var orders []models.MembershipOrder
	query := c.DB.Where("status = ?", models.OrderStatusPending).Order("created_at DESC")

	// 关联查询用户信息
	query = query.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, email, phone")
	})

	if err := query.Find(&orders).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取待审核订单失败"})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// ApproveOrder 管理员审核通过订单
func (c *MembershipController) ApproveOrder(ctx *gin.Context) {
	// 验证是否是管理员
	roleInterface, exists := ctx.Get("role")
	if !exists || roleInterface.(string) != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	adminIDInterface, _ := ctx.Get("user_id")
	adminID := adminIDInterface.(uint)

	orderID := ctx.Param("id")

	// 使用事务保证原子性
	tx := c.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询订单
	var order models.MembershipOrder
	if err := tx.First(&order, orderID).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "订单不存在"})
		return
	}

	// 检查订单状态
	if order.Status != models.OrderStatusPending {
		tx.Rollback()
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不正确，无法审核"})
		return
	}

	// 更新订单状态
	order.Status = models.OrderStatusApproved
	order.AdminID = adminID
	order.ApprovedAt = time.Now()

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新订单状态失败"})
		return
	}

	// 更新用户会员
	var membership models.Membership
	result := tx.Where("user_id = ?", order.UserID).First(&membership)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员信息失败"})
		return
	}

	startDate := time.Now()
	var expireDate time.Time
	if order.Duration > 0 {
		expireDate = startDate.AddDate(0, 0, order.Duration)
	} else {
		expireDate = time.Date(9999, 12, 31, 23, 59, 59, 0, time.Local) // 永不过期
	}

	// 如果存在会员记录且未过期，则延长有效期
	if result.Error == nil && membership.Level != models.MembershipFree && time.Now().Before(membership.ExpireDate) {
		expireDate = membership.ExpireDate.AddDate(0, 0, order.Duration)
	}

	// 查询会员计划获取任务优先级和每日限制
	var plan models.MembershipPlan
	if err := tx.First(&plan, order.PlanID).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员计划失败"})
		return
	}

	if result.Error == gorm.ErrRecordNotFound {
		// 创建新会员记录
		membership = models.Membership{
			UserID:       order.UserID,
			Level:        order.Level,
			StartDate:    startDate,
			ExpireDate:   expireDate,
			AutoRenew:    order.AutoRenew,
			PaymentID:    orderID,
			Status:       "active",
			TaskPriority: plan.TaskPriority,
			DailyLimit:   plan.DailyLimit,
		}
		if err := tx.Create(&membership).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建会员记录失败"})
			return
		}
	} else {
		// 更新现有会员记录
		membership.Level = order.Level
		membership.StartDate = startDate
		membership.ExpireDate = expireDate
		membership.AutoRenew = order.AutoRenew
		membership.PaymentID = orderID
		membership.Status = "active"
		membership.TaskPriority = plan.TaskPriority
		membership.DailyLimit = plan.DailyLimit

		if err := tx.Save(&membership).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新会员记录失败"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "订单审核通过，会员已激活",
		"order":   order,
	})
}

// RejectOrder 管理员拒绝订单
func (c *MembershipController) RejectOrder(ctx *gin.Context) {
	// 验证是否是管理员
	roleInterface, exists := ctx.Get("role")
	if !exists || roleInterface.(string) != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	adminIDInterface, _ := ctx.Get("user_id")
	adminID := adminIDInterface.(uint)

	orderID := ctx.Param("id")

	var req struct {
		RejectReason string `json:"reject_reason"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查询订单
	var order models.MembershipOrder
	if err := c.DB.First(&order, orderID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "订单不存在"})
		return
	}

	// 检查订单状态
	if order.Status != models.OrderStatusPending {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不正确，无法拒绝"})
		return
	}

	// 更新订单状态
	order.Status = models.OrderStatusRejected
	order.AdminID = adminID
	order.RejectedAt = time.Now()
	order.RejectReason = req.RejectReason

	if err := c.DB.Save(&order).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新订单状态失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "订单已拒绝",
		"order":   order,
	})
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
