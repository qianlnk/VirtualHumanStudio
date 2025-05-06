package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/services"
)

var membershipServiceInstance services.MembershipService

// SetMembershipService 设置会员服务实例
func SetMembershipService(service services.MembershipService) {
	membershipServiceInstance = service
}

// MembershipCheck 会员检查中间件
func MembershipCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if membershipServiceInstance == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "会员服务未初始化"})
			c.Abort()
			return
		}

		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要登录"})
			c.Abort()
			return
		}

		// 检查用户是否超出使用限制
		allowed, err := membershipServiceInstance.CheckUsageLimit(userID.(uint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查使用限制失败"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "您今日的使用次数已达上限，请升级会员或明日再试",
				"code":  "usage_limit_exceeded",
			})
			c.Abort()
			return
		}

		// 增加使用量
		if err := membershipServiceInstance.IncrementUsage(userID.(uint)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "增加使用量失败"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// MembershipAccess 检查会员访问级别
func MembershipAccess(minLevel string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if membershipServiceInstance == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "会员服务未初始化"})
			c.Abort()
			return
		}

		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要登录"})
			c.Abort()
			return
		}

		// 获取用户会员信息
		membership, err := membershipServiceInstance.GetMembershipInfo(userID.(uint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员信息失败"})
			c.Abort()
			return
		}

		// 检查会员级别
		if membership.Level == "free" && minLevel != "free" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "此功能需要会员才能使用",
				"code":  "membership_required",
			})
			c.Abort()
			return
		}

		// 检查会员状态
		if membership.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "您的会员已过期，请续费",
				"code":  "membership_expired",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
