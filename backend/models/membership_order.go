package models

import (
	"time"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending  OrderStatus = "pending"  // 待审核
	OrderStatusApproved OrderStatus = "approved" // 已审核
	OrderStatusRejected OrderStatus = "rejected" // 已拒绝
	OrderStatusExpired  OrderStatus = "expired"  // 已过期
)

// PaymentMethod 支付方式
type PaymentMethod string

const (
	PaymentMethodWechat PaymentMethod = "wechat" // 微信支付
	PaymentMethodAlipay PaymentMethod = "alipay" // 支付宝支付
)

// MembershipOrder 会员订单模型
type MembershipOrder struct {
	BaseModel
	UserID        uint            `json:"user_id" gorm:"index;not null"`           // 用户ID
	User          User            `json:"user" gorm:"foreignKey:UserID"`           // 用户信息
	PlanID        uint            `json:"plan_id" gorm:"index;not null"`           // 会员计划ID
	PlanName      string          `json:"plan_name" gorm:"size:50;not null"`       // 会员计划名称
	Level         MembershipLevel `json:"level" gorm:"size:20;not null"`           // 会员等级
	Price         float64         `json:"price"`                                   // 价格
	Duration      int             `json:"duration"`                                // 持续时间(天)
	Status        OrderStatus     `json:"status" gorm:"size:20;default:'pending'"` // 订单状态
	PaymentMethod PaymentMethod   `json:"payment_method" gorm:"size:20"`           // 支付方式
	PaymentRemark string          `json:"payment_remark" gorm:"size:200"`          // 支付备注
	AutoRenew     bool            `json:"auto_renew" gorm:"default:false"`         // 是否自动续费
	AdminID       uint            `json:"admin_id"`                                // 处理订单的管理员ID
	ApprovedAt    time.Time       `json:"approved_at"`                             // 审核通过时间
	RejectedAt    time.Time       `json:"rejected_at"`                             // 拒绝时间
	RejectReason  string          `json:"reject_reason" gorm:"size:200"`           // 拒绝原因
}
