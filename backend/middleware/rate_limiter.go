package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	redis_rate "github.com/go-redis/redis_rate/v9"
)

// 初始化Redis限流器
var limiter *redis_rate.Limiter

func InitRateLimiter() {
	// 创建限流器
	limiter = redis_rate.NewLimiter(redisClient)
}

func RateLimiter(featureType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 用户id
		userId := c.GetString("user_id")
		key := fmt.Sprintf("user_id:%s_feature_type:%s", userId, featureType)

		res, err := limiter.Allow(c, key, redis_rate.PerSecond(1))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiting failed"})
			c.Abort()
			return
		}
		if res.Allowed == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
