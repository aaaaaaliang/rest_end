package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config 结构体映射 config.yaml
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

// G 全局配置
var G Config

// LoadConfig 读取配置文件
func LoadConfig() {
	viper.SetConfigName("config") // 读取 config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config") // 指定目录

	// 读取环境变量（优先级高于配置文件）
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("❌ 读取配置失败: %v", err)
	}

	// 解析到全局变量
	if err := viper.Unmarshal(&G); err != nil {
		log.Fatalf("❌ 配置解析失败: %v", err)
	}

	fmt.Println("✔️ 配置加载成功！")
}
