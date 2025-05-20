package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/services"
)

var membershipServiceInstance services.MembershipService

// SetMembershipService 设置会员服务实例
func SetMembershipService(service services.MembershipService) {
	membershipServiceInstance = service
}

// FeatureUsageCheck 特定功能使用限制检查中间件
func FeatureUsageCheck(featureType models.FeatureType) gin.HandlerFunc {
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

		// 检查特定功能的使用限制
		allowed, errMsg, err := membershipServiceInstance.CheckFeatureUsageLimit(userID.(uint), featureType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查功能使用限制失败"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errMsg,
				"code":  "feature_usage_limit_exceeded",
			})
			c.Abort()
			return
		}

		// 设置功能类型到上下文，供后续使用
		c.Set("feature_type", featureType)

		c.Next()
	}
}

// IncrementFeatureUsage 增加特定功能使用量的中间件
func IncrementFeatureUsage(usageValue int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续的处理
		c.Next()

		// 请求完成后，记录功能使用量
		userID, exists := c.Get("user_id")
		if !exists {
			return
		}

		featureTypeInterface, exists := c.Get("feature_type")
		if !exists {
			return
		}

		featureType, ok := featureTypeInterface.(models.FeatureType)
		if !ok {
			return
		}

		ctxUsageValue, exexists := c.Get("usage_value")
		if exexists {
			usageValue = ctxUsageValue.(int)
		}

		// 增加特定功能的使用量
		_ = membershipServiceInstance.IncrementFeatureUsage(userID.(uint), featureType, usageValue)
	}
}

// VoiceCloneCheck 音色克隆功能检查中间件
func VoiceCloneCheck() gin.HandlerFunc {
	return FeatureUsageCheck(models.FeatureVoiceClone)
}

// TTSCheck 语音合成功能检查中间件
func TTSCheck() gin.HandlerFunc {
	return FeatureUsageCheck(models.FeatureTTS)
}

// ASRCheck 语音识别功能检查中间件
func ASRCheck() gin.HandlerFunc {
	return FeatureUsageCheck(models.FeatureASR)
}

// DigitalHumanCheck 数字人合成功能检查中间件
func DigitalHumanCheck() gin.HandlerFunc {
	return FeatureUsageCheck(models.FeatureDigitalHuman)
}

// ImageProcessCheck 图片处理功能检查中间件
func ImageProcessCheck() gin.HandlerFunc {
	return FeatureUsageCheck(models.FeatureImageProcess)
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
