package models

import (
	"time"
)

// UserUsage 用户使用量模型
type UserUsage struct {
	UserID     uint      `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	Date       time.Time `json:"date" gorm:"primaryKey;type:date"`
	UsageCount int       `json:"usage_count" gorm:"default:0"`
}

// TableName 设置表名
func (UserUsage) TableName() string {
	return "user_usages"
}
