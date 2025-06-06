package models

import "time"

// Message 留言模型
type Message struct {
	BaseModel
	UserID    uint      `json:"user_id" gorm:"index;not null"`          // 用户ID
	Name      string    `json:"name" gorm:"size:100;not null"`          // 留言者姓名
	Email     string    `json:"email" gorm:"size:100;not null"`         // 联系邮箱
	Phone     string    `json:"phone" gorm:"size:20"`                   // 联系电话
	Subject   string    `json:"subject" gorm:"size:200;not null"`       // 留言主题
	Content   string    `json:"content" gorm:"type:text;not null"`      // 留言内容
	Status    string    `json:"status" gorm:"size:20;default:'unread'"` // 状态: unread, read, replied
	ReplyText string    `json:"reply_text" gorm:"type:text"`            // 回复内容
	ReplyTime time.Time `json:"reply_time"`                             // 回复时间
	AdminID   uint      `json:"admin_id"`                               // 回复的管理员ID
}
