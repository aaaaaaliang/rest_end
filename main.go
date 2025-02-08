package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/xormplus/xorm"
	"log"
	"rest/config"
	"rest/model"
	"rest/route"
)

func main() {
	// **加载配置**
	config.LoadConfig()
	// **初始化配置**
	config.InitConfig()

	if err := config.CreateTables(
		&model.Category{},
	); err != nil {
		log.Fatal(err)
	}

	// **创建 Gin 实例**
	r := gin.Default()

	// **注册路由**
	route.RegisterRoutes(r)

	// **获取端口 & 启动服务**
	host := config.G.App.Host
	port := config.G.App.Port
	fmt.Printf("🚀 服务启动: %v:%d\n", host, port)
	log.Fatal(r.Run(fmt.Sprintf(":%d", port)))
}
