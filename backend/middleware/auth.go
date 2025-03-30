package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"VirtualHumanStudio/backend/config"
	"VirtualHumanStudio/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
)

// Redis客户端
var redisClient *redis.Client

// InitRedis 初始化Redis客户端
func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
		Password: config.AppConfig.RedisPassword,
		DB:       config.AppConfig.RedisDB,
	})

	// 测试连接
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
	} else {
		fmt.Println("Connected to Redis successfully")
	}
}

// GetRedisClient 获取Redis客户端
func GetRedisClient() *redis.Client {
	return redisClient
}

// JWTClaims 自定义JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(user models.User) (string, error) {
	// 设置JWT声明
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 令牌有效期24小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "virtual_human_studio",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	// 将令牌存储在Redis中，用于注销功能
	ctx := context.Background()
	key := fmt.Sprintf("token:%d", user.ID)
	err = redisClient.Set(ctx, key, tokenString, 24*time.Hour).Err()
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*JWTClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 提取令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式无效"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析令牌
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		// 检查令牌是否已被注销
		ctx := context.Background()
		key := fmt.Sprintf("token:%d", claims.UserID)
		val, err := redisClient.Get(ctx, key).Result()
		if err != nil || val != tokenString {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌已失效"})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行JWT认证
		JWTAuth()(c)

		// 如果认证失败，中间件链会中断
		if c.IsAborted() {
			return
		}

		// 检查用户角色
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Logout 注销用户
func Logout(userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("token:%d", userID)
	return redisClient.Del(ctx, key).Err()
}
