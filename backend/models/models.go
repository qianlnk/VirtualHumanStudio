package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// MembershipLevel 会员等级
type MembershipLevel string

const (
	MembershipFree    MembershipLevel = "free"    // 免费用户
	MembershipMonthly MembershipLevel = "monthly" // 月会员
	MembershipQuarter MembershipLevel = "quarter" // 季会员
	MembershipYearly  MembershipLevel = "yearly"  // 年会员
)

// Membership 会员模型
type Membership struct {
	BaseModel
	UserID       uint            `json:"user_id" gorm:"uniqueIndex;not null"`
	Level        MembershipLevel `json:"level" gorm:"size:20;default:'free'"`
	StartDate    time.Time       `json:"start_date"`                             // 会员开始时间
	ExpireDate   time.Time       `json:"expire_date"`                            // 会员过期时间
	AutoRenew    bool            `json:"auto_renew" gorm:"default:false"`        // 是否自动续费
	PaymentID    string          `json:"payment_id" gorm:"size:100"`             // 支付ID
	Status       string          `json:"status" gorm:"size:20;default:'active'"` // active, expired
	TaskPriority int             `json:"task_priority" gorm:"default:0"`         // 任务优先级，数字越大优先级越高
	DailyLimit   int             `json:"daily_limit"`                            // 每日使用次数限制，-1表示无限制
}

// User 用户模型
type User struct {
	BaseModel
	Username      string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password      string         `json:"-" gorm:"size:100;not null"`
	Email         string         `json:"email" gorm:"size:100"`
	Phone         string         `json:"phone" gorm:"size:20"`
	Role          string         `json:"role" gorm:"size:20;default:'user'"` // admin, user
	Status        int            `json:"status" gorm:"default:1"`            // 0: 禁用, 1: 正常
	LastLoginAt   time.Time      `json:"last_login_at"`
	LastLoginIP   string         `json:"last_login_ip" gorm:"size:50"`
	Membership    Membership     `json:"membership,omitempty" gorm:"foreignKey:UserID"`
	VoiceClones   []VoiceClone   `json:"voice_clones,omitempty" gorm:"foreignKey:UserID"`
	TTSTasks      []TTSTask      `json:"tts_tasks,omitempty" gorm:"foreignKey:UserID"`
	DigitalHumans []DigitalHuman `json:"digital_humans,omitempty" gorm:"foreignKey:UserID"`
}

// VoiceClone 音色克隆模型
type VoiceClone struct {
	BaseModel
	UserID      uint   `json:"user_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	ModelName   string `json:"model_name" gorm:"size:100;not null"`
	PromptFile  string `json:"prompt_file" gorm:"size:255;not null"` // 文件路径
	PromptText  string `json:"prompt_text" gorm:"size:1000"`
	SpeakerName string `json:"speaker_name" gorm:"size:100;not null"`
	Status      string `json:"status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	Result      string `json:"result" gorm:"size:255"`                  // 结果文件路径
	TaskID      string `json:"task_id" gorm:"size:100"`                 // 外部API任务ID
	ErrorMsg    string `json:"error_msg" gorm:"size:500"`
	SampleFile  string `json:"sample_file" gorm:"size:255"` // 试听音频文件路径
}

// TTSTask 文本转语音任务模型
type TTSTask struct {
	BaseModel
	UserID      uint   `json:"user_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	InputText   string `json:"input_text" gorm:"type:text"`               // 输入文本
	OutputFile  string `json:"output_file" gorm:"size:255"`               // 输出音频文件路径
	SpeakerName string `json:"speaker_name" gorm:"size:100"`              // 使用的音色名称
	Type        string `json:"type" gorm:"size:20;default:'text2speech'"` // 任务类型，默认为text
	Status      string `json:"status" gorm:"size:20;default:'pending'"`   // pending, processing, completed, failed
	TaskID      string `json:"task_id" gorm:"size:100"`                   // 外部API任务ID
	ErrorMsg    string `json:"error_msg" gorm:"size:500"`
}

// DigitalHuman 数字人模型
type DigitalHuman struct {
	BaseModel
	UserID          uint   `json:"user_id" gorm:"index;not null"`
	Name            string `json:"name" gorm:"size:100;not null"`
	Description     string `json:"description" gorm:"size:500"`
	AudioURL        string `json:"audio_url" gorm:"size:255;not null"`            // 音频文件路径
	VideoURL        string `json:"video_url" gorm:"size:255;not null"`            // 视频文件路径
	TaskCode        string `json:"task_code" gorm:"size:100;not null"`            // 任务代码
	Chaofen         int    `json:"chaofen" gorm:"default:0"`                      // 超分参数
	WatermarkSwitch int    `json:"watermark_switch" gorm:"default:0"`             // 水印开关
	PN              int    `json:"pn" gorm:"default:1"`                           // PN参数
	Status          string `json:"status" gorm:"size:20;default:'pending'"`       // pending, processing, completed, failed
	ResultURL       string `json:"result_url" gorm:"size:255"`                    // 结果文件URL
	ErrorMsg        string `json:"error_msg" gorm:"size:500"`                     //
	ShareStatus     string `json:"share_status" gorm:"size:20;default:'private'"` // private, pending_review, approved, rejected
}

type ShareTask struct {
	ID           uint       `json:"id" gorm:"primaryKey"`                          // 主键ID
	TaskID       uint       `json:"task_id" gorm:"index;not null"`                 // 任务ID
	Mode         string     `json:"mode" gorm:"size:20;not null"`                  // comfyui, digital_human
	TaskType     string     `json:"task_type" gorm:"size:20;not null"`             // 任务类型
	ShareStatus  string     `json:"share_status" gorm:"size:20;default:'private'"` // private, pending_review, approved, rejected
	ShareTime    *time.Time `json:"share_time"`                                    // 分享时间
	ReviewerID   *uint      `json:"reviewer_id" gorm:"index"`                      // 审核人ID
	ReviewTime   *time.Time `json:"review_time"`                                   // 审核时间
	RejectReason string     `json:"reject_reason" gorm:"size:200"`                 // 拒绝原因
}

// 点赞
type ShareTaskLike struct {
	BaseModel
	UserID      uint `json:"user_id" gorm:"index;not null"`       // 用户ID
	ShareTaskID uint `json:"share_task_id" gorm:"index;not null"` // 分享任务ID
}

// 收藏
type ShareTaskFavorite struct {
	BaseModel
	UserID      uint `json:"user_id" gorm:"index;not null"`       // 用户ID
	ShareTaskID uint `json:"share_task_id" gorm:"index;not null"` // 分享任务ID
}

// 评论
type ShareTaskComment struct {
	BaseModel
	UserID      uint   `json:"user_id" gorm:"index;not null"`       // 用户ID
	ShareTaskID uint   `json:"share_task_id" gorm:"index;not null"` // 分享任务ID
	Content     string `json:"content" gorm:"size:500"`             // 评论内容
}

// VoiceLibrary 音色库模型
type VoiceLibrary struct {
	BaseModel
	Name        string `json:"name" gorm:"size:100;not null"`       // 音色名称
	Description string `json:"description" gorm:"size:500"`         // 音色描述
	ModelName   string `json:"model_name" gorm:"size:100;not null"` // 模型名称
	ModelFile   string `json:"model_file" gorm:"size:255;not null"` // 音色模型文件路径
	SampleFile  string `json:"sample_file" gorm:"size:255"`         // 试听音频文件路径
	Type        string `json:"type" gorm:"size:20;not null"`        // original, cloned
	OwnerID     uint   `json:"owner_id" gorm:"index"`               // 所有者ID，如果是克隆音色
	IsPublic    bool   `json:"is_public" gorm:"default:false"`      // 是否公开
}

// MembershipPlan 会员计划模型
type MembershipPlan struct {
	BaseModel
	Name         string          `json:"name" gorm:"size:50;not null"`   // 计划名称
	Level        MembershipLevel `json:"level" gorm:"size:20;not null"`  // 会员等级
	Price        float64         `json:"price"`                          // 价格
	Duration     int             `json:"duration"`                       // 持续时间(天)
	TaskPriority int             `json:"task_priority" gorm:"default:0"` // 任务优先级
	DailyLimit   int             `json:"daily_limit"`                    // 每日使用次数限制
	Description  string          `json:"description" gorm:"size:500"`    // 描述
	Features     string          `json:"features" gorm:"type:text"`      // 功能特性(JSON格式)
	IsActive     bool            `json:"is_active" gorm:"default:true"`  // 是否激活
	WechatQRCode string          `json:"wechat_qr_code" gorm:"size:255"` // 微信支付二维码图片URL
	AlipayQRCode string          `json:"alipay_qr_code" gorm:"size:255"` // 支付宝支付二维码图片URL
}

// 简洁的用户信息
type UserInfo struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}
