package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qianlnk/VirtualHumanStudio/backend/services"
	"github.com/qianlnk/VirtualHumanStudio/backend/utils"

	"github.com/gin-gonic/gin"
)

// GetUserStatistics 获取用户统计数据
func GetUserStatistics(c *gin.Context) {
	// 验证管理员权限
	role, exists := c.Get("role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 获取日期范围参数，默认为最近7天
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效"})
		return
	}
	endDate = endDate.Add(24 * time.Hour) // 包含结束日期当天

	// 获取日期范围内的统计数据
	stats, err := services.GetDailyStatistics(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计数据失败"})
		return
	}

	// 返回统计数据
	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// GetModuleUsageStatistics 获取模块使用统计数据
func GetModuleUsageStatistics(c *gin.Context) {
	// 验证管理员权限
	role, exists := c.Get("role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 获取日期范围参数，默认为最近7天
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效"})
		return
	}
	endDate = endDate.Add(24 * time.Hour) // 包含结束日期当天

	// 获取模块使用统计
	moduleUsage, err := services.GetModuleUsage(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模块使用统计失败"})
		return
	}

	// 获取用户活跃度排名（前10名）
	userActivity, err := services.GetUserActivity(startDate, endDate, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户活跃度排名失败"})
		return
	}

	// 返回统计数据
	c.JSON(http.StatusOK, gin.H{
		"module_usage":  moduleUsage,
		"user_activity": userActivity,
		"start_date":    startDate.Format("2006-01-02"),
		"end_date":      endDateStr,
	})
}

// GetUserLoginLogs 获取用户登录日志
func GetUserLoginLogs(c *gin.Context) {
	// 验证管理员权限
	role, exists := c.Get("role")
	if !exists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 分页参数
	page, size := utils.GetPaginationParams(c)

	// 获取日期范围参数，默认为最近7天
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效"})
		return
	}
	endDate = endDate.Add(24 * time.Hour) // 包含结束日期当天

	// 可选的用户ID过滤
	var userID uint = 0
	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		_, err := fmt.Sscanf(userIDStr, "%d", &userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式无效"})
			return
		}
	}

	// 获取登录日志
	logs, count, err := services.GetUserLoginLogs(startDate, endDate, userID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取登录日志失败"})
		return
	}

	// 返回登录日志
	c.JSON(http.StatusOK, gin.H{
		"total": count,
		"page":  page,
		"size":  size,
		"logs":  logs,
	})
}
