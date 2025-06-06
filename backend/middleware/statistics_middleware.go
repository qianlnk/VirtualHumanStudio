package middleware

import (
	"fmt"
	"strings"

	"github.com/qianlnk/VirtualHumanStudio/backend/services"

	"github.com/gin-gonic/gin"
)

// 模块路径映射表
var modulePathMap = map[string]string{
	"/api/voice/clone":      "音色克隆",
	"/api/voice/clones":     "音色克隆",
	"/api/voice":            "音色库",
	"/api/voices":           "音色库",
	"/api/tts":              "文本转语音",
	"/api/asr":              "语音识别",
	"/api/digital-human":    "数字人",
	"/api/accessory":        "饰品替换",
	"/api/image-processing": "图像处理",
	"/api/message":          "留言",
	"/api/messages":         "留言",
}

// StatisticsMiddleware 统计中间件，用于记录用户的模块使用情况
func StatisticsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续的处理
		c.Next()

		// 请求完成后，记录用户访问统计
		// 只有已认证的用户才记录统计
		userID, exists := c.Get("user_id")
		if !exists {
			return
		}

		username, _ := c.Get("username")
		usernameStr, ok := username.(string)
		if !ok {
			usernameStr = "unknown"
		}

		// 获取请求路径
		path := c.Request.URL.Path

		// 查找对应的模块名称
		moduleName := "其他"
		for prefix, name := range modulePathMap {
			if strings.HasPrefix(path, prefix) {
				moduleName = name
				break
			}
		}

		// 获取请求方法作为操作类型
		actionType := c.Request.Method

		// 尝试获取资源ID
		resourceID := uint(0)
		// 解析路径中的ID参数
		parts := strings.Split(path, "/")
		if len(parts) > 3 {
			// 尝试将最后一部分解析为ID
			lastPart := parts[len(parts)-1]
			var id uint
			_, err := fmt.Sscanf(lastPart, "%d", &id)
			if err == nil {
				resourceID = id
			}
		}

		// 记录模块使用统计
		go services.RecordModuleUsage(userID.(uint), usernameStr, moduleName, actionType, resourceID, path)
	}
}
