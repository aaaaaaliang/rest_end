package config

import (
	"fmt"
	"github.com/xormplus/xorm"
	"log"
	"sync"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	LogLevel     string `mapstructure:"log_level"`
	ShowSQL      bool   `mapstructure:"show_sql"`
}

// DB **全局数据库连接**
var DB *xorm.Engine

// InitDB **初始化 xorm 连接**
func InitDB() {
	dbConfig := G.Database

	// 拼接 MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)

	// 创建数据库引擎
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 设置连接池参数
	engine.SetMaxOpenConns(dbConfig.MaxOpenConns)
	engine.SetMaxIdleConns(dbConfig.MaxIdleConns)

	// 设置日志级别
	if dbConfig.ShowSQL {
		engine.ShowSQL(true)
	} else {
		engine.ShowSQL(false)
	}

	// **Ping 数据库，确保连接可用**
	if err := engine.Ping(); err != nil {
		//logger.Fatalf("❌ 数据库连接测试失败: %v", err)
		log.Printf("❌ 数据库连接测试失败: %v", err)
	}

	DB = engine
	log.Println("✔️ 数据库连接成功！")
}

// CreateTables 创建表
func CreateTables(models ...interface{}) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(models)) // 错误通道

	for _, model := range models {
		wg.Add(1)
		go func(m interface{}) {
			defer wg.Done()
			if err := DB.Sync(m); err != nil {
				errCh <- fmt.Errorf("同步表失败 %T: %v", m, err)
			}
		}(model)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}
	return nil
}
