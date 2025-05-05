package models

import "time"

// UserLoginLog 用户登录日志
type UserLoginLog struct {
	BaseModel
	UserID    uint      `json:"user_id" gorm:"index;not null"` // 用户ID
	Username  string    `json:"username" gorm:"size:50"`       // 用户名
	LoginTime time.Time `json:"login_time"`                    // 登录时间
	LoginIP   string    `json:"login_ip" gorm:"size:50"`       // 登录IP
}

// ModuleUsageLog 模块使用日志
type ModuleUsageLog struct {
	BaseModel
	UserID      uint      `json:"user_id" gorm:"index;not null"`    // 用户ID
	Username    string    `json:"username" gorm:"size:50"`          // 用户名
	ModuleName  string    `json:"module_name" gorm:"size:50;index"` // 模块名称
	ActionType  string    `json:"action_type" gorm:"size:30"`       // 操作类型 (例如: create, view, delete)
	ResourceID  uint      `json:"resource_id"`                      // 资源ID
	AccessTime  time.Time `json:"access_time"`                      // 访问时间
	RequestPath string    `json:"request_path" gorm:"size:255"`     // 请求路径
}

// DailyStatistics 每日统计数据
type DailyStatistics struct {
	BaseModel
	Date           time.Time `json:"date" gorm:"uniqueIndex;type:date"` // 日期
	ActiveUsers    int       `json:"active_users"`                      // 活跃用户数 (DAU)
	NewUsers       int       `json:"new_users"`                         // 新增用户数
	TotalUsers     int       `json:"total_users"`                       // 总用户数
	TotalLogins    int       `json:"total_logins"`                      // 总登录次数
	TotalRequests  int       `json:"total_requests"`                    // 总请求数
	ModuleUsageMap string    `json:"module_usage_map" gorm:"type:text"` // 模块使用情况 (JSON格式)
}
