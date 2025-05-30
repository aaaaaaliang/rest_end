package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config 结构体映射 config.production.yaml
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	//Log      LogConfig      `mapstructure:"logger"`
	Oauth2  Oauth2Config `mapstructure:"oauth2"` // Oauth2的结构体
	Cors    CorsConfig   `mapstructure:"cors"`   // CORS 配置
	Uploads UploadConfig `mapstructure:"uploads"`
	AI      AIConfig     `mapstructure:"ai"`
	Pay     PayConfig    `mapstructure:"pay"`
	MQ      MqConfig     `mapstructure:"mq"`
	ES      ESConfig     `mapstructure:"es"`
}

// G 全局配置
var G Config

// LoadConfig 读取配置文件
func LoadConfig() {

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	viper.SetConfigName("config." + env) // 读取 config.production.yaml
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
