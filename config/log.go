package config

import (
	"log"
	"os"
)

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// InitLogger **初始化日志系统**
func InitLogger() {
	logConfig := G.Log

	// 创建日志文件
	logFile, err := os.OpenFile(logConfig.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("❌ 日志文件创建失败: %v", err)
	}

	// 设置日志输出到文件
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("✔️ 日志系统初始化成功")
}
