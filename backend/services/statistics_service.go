package services

import (
	"encoding/json"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
)

// RecordModuleUsage 记录模块使用统计
func RecordModuleUsage(userID uint, username string, moduleName string, actionType string, resourceID uint, requestPath string) error {
	usage := models.ModuleUsageLog{
		UserID:      userID,
		Username:    username,
		ModuleName:  moduleName,
		ActionType:  actionType,
		ResourceID:  resourceID,
		AccessTime:  time.Now(),
		RequestPath: requestPath,
	}

	result := db.DB.Create(&usage)
	return result.Error
}

// RecordUserLogin 记录用户登录统计
func RecordUserLogin(userID uint, username string, ip string) error {
	loginLog := models.UserLoginLog{
		UserID:    userID,
		Username:  username,
		LoginTime: time.Now(),
		LoginIP:   ip,
	}

	result := db.DB.Create(&loginLog)
	return result.Error
}

// UpdateDailyStatistics 更新每日统计数据
func UpdateDailyStatistics() error {
	today := time.Now().Truncate(24 * time.Hour)

	var stat models.DailyStatistics
	result := db.DB.Where("date = ?", today).First(&stat)

	// 如果今日统计不存在，创建一个新的记录
	if result.Error != nil && result.Error.Error() == "record not found" {
		// 计算总用户数
		var totalUsers int64
		db.DB.Model(&models.User{}).Count(&totalUsers)

		// 计算今日新增用户数
		var newUsers int64
		db.DB.Model(&models.User{}).Where("created_at >= ?", today).Count(&newUsers)

		// 计算今日活跃用户数
		var activeUsers int64
		db.DB.Model(&models.UserLoginLog{}).Where("login_time >= ?", today).Distinct("user_id").Count(&activeUsers)

		// 计算今日登录次数
		var totalLogins int64
		db.DB.Model(&models.UserLoginLog{}).Where("login_time >= ?", today).Count(&totalLogins)

		// 计算今日总请求数
		var totalRequests int64
		db.DB.Model(&models.ModuleUsageLog{}).Where("access_time >= ?", today).Count(&totalRequests)

		// 计算各模块使用情况
		var moduleUsage []struct {
			ModuleName string
			Count      int
		}
		db.DB.Model(&models.ModuleUsageLog{}).
			Select("module_name, COUNT(*) as count").
			Where("access_time >= ?", today).
			Group("module_name").
			Scan(&moduleUsage)

		// 序列化模块使用情况为JSON
		moduleUsageJSON, _ := json.Marshal(moduleUsage)

		// 创建今日统计记录
		stat = models.DailyStatistics{
			Date:           today,
			ActiveUsers:    int(activeUsers),
			NewUsers:       int(newUsers),
			TotalUsers:     int(totalUsers),
			TotalLogins:    int(totalLogins),
			TotalRequests:  int(totalRequests),
			ModuleUsageMap: string(moduleUsageJSON),
		}

		result = db.DB.Create(&stat)
	} else if result.Error == nil {
		// 更新今日统计记录
		var activeUsers int64
		db.DB.Model(&models.UserLoginLog{}).Where("login_time >= ?", today).Distinct("user_id").Count(&activeUsers)

		var totalLogins int64
		db.DB.Model(&models.UserLoginLog{}).Where("login_time >= ?", today).Count(&totalLogins)

		var totalRequests int64
		db.DB.Model(&models.ModuleUsageLog{}).Where("access_time >= ?", today).Count(&totalRequests)

		var moduleUsage []struct {
			ModuleName string
			Count      int
		}
		db.DB.Model(&models.ModuleUsageLog{}).
			Select("module_name, COUNT(*) as count").
			Where("access_time >= ?", today).
			Group("module_name").
			Scan(&moduleUsage)

		moduleUsageJSON, _ := json.Marshal(moduleUsage)

		stat.ActiveUsers = int(activeUsers)
		stat.TotalLogins = int(totalLogins)
		stat.TotalRequests = int(totalRequests)
		stat.ModuleUsageMap = string(moduleUsageJSON)

		result = db.DB.Save(&stat)
	}

	return result.Error
}

// GetDailyStatistics 获取指定日期范围的每日统计数据
func GetDailyStatistics(startDate, endDate time.Time) ([]models.DailyStatistics, error) {
	var stats []models.DailyStatistics
	result := db.DB.Where("date >= ? AND date < ?", startDate, endDate).Order("date").Find(&stats)
	return stats, result.Error
}

// GetModuleUsage 获取指定日期范围的模块使用统计
func GetModuleUsage(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var moduleUsage []map[string]interface{}

	rows, err := db.DB.Model(&models.ModuleUsageLog{}).
		Select("module_name, COUNT(*) as count").
		Where("access_time >= ? AND access_time < ?", startDate, endDate).
		Group("module_name").
		Order("count DESC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var moduleName string
		var count int
		rows.Scan(&moduleName, &count)

		moduleUsage = append(moduleUsage, map[string]interface{}{
			"module_name": moduleName,
			"count":       count,
		})
	}

	return moduleUsage, nil
}

// GetUserActivity 获取指定日期范围的用户活跃度排名
func GetUserActivity(startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	var userActivity []map[string]interface{}

	rows, err := db.DB.Model(&models.ModuleUsageLog{}).
		Select("user_id, username, COUNT(*) as count").
		Where("access_time >= ? AND access_time < ?", startDate, endDate).
		Group("user_id, username").
		Order("count DESC").
		Limit(limit).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID uint
		var username string
		var count int
		rows.Scan(&userID, &username, &count)

		userActivity = append(userActivity, map[string]interface{}{
			"user_id":  userID,
			"username": username,
			"count":    count,
		})
	}

	return userActivity, nil
}

// GetUserLoginLogs 获取用户登录日志
func GetUserLoginLogs(startDate, endDate time.Time, userID uint, page, size int) ([]models.UserLoginLog, int64, error) {
	// 构建查询
	query := db.DB.Model(&models.UserLoginLog{}).
		Where("login_time >= ? AND login_time < ?", startDate, endDate)

	// 如果指定了用户ID，添加到查询条件
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 查询总数
	var count int64
	query.Count(&count)

	// 获取登录日志
	var logs []models.UserLoginLog
	result := query.Order("login_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&logs)

	return logs, count, result.Error
}
