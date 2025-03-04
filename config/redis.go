package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
)

// RedisConfig 配置结构体
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

// R **全局 Redis 客户端**
var R *redis.Client

// InitRedis **初始化 Redis 连接**
func InitRedis() {
	redisConfig := G.Redis

	// 连接 Redis
	R = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password, // 没有密码则留空
		DB:       redisConfig.DB,
	})

	// **测试 Redis 连接**
	ctx := context.Background()
	_, err := R.Ping(ctx).Result()
	if err != nil {
		log.Printf("❌ Redis 连接失败: %v", err)
	}

	log.Println("✔️ Redis 连接成功！")
}
