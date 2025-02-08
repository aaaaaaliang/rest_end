package config

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"log"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

// R **全局 Redis 连接**
var R *redis.Client
var ctx = context.Background()

// InitRedis **初始化 Redis 连接**
func InitRedis() {
	redisConfig := G.Redis

	// 连接 Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password, // 密码
		DB:       redisConfig.DB,       // 数据库
	})

	// Ping Redis，确保连接可用
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis 连接失败: %v", err)
	}

	R = client
	log.Println("✔️ Redis 连接成功！")
}
