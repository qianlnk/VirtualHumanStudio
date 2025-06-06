package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/qianlnk/VirtualHumanStudio/backend/config"
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
