package main

import (
	_ "github.com/go-sql-driver/mysql"
	"log"
	"rest/api/coupon"
	"rest/api/order"
	"rest/config"
	"rest/logger"
	"rest/model"
	"rest/route"
)

func main() {
	// **加载配置**
	config.LoadConfig()
	// **初始化配置**
	config.InitConfig()
	logger.InitLogger()
	config.InitJWT()

	// **创建数据库表**
	if err := config.CreateTables(
		&model.Category{},
		&model.Users{},
		&model.Products{},
		&model.UserCart{},
		&model.UserOrder{},
		&model.OrderDetail{},
		&model.Banner{},
		&model.APIPermission{},
		&model.Role{},
		&model.RolePermission{},
		&model.UserRole{},
		&model.SalaryRecord{},
		&model.ChatMessage{},
		&model.TableInfo{},
		&model.CouponTemplate{},
		&model.UserCoupon{},
		&model.AIModelConfig{},
		&model.AIChatHistory{},
		&model.AIChatSession{},
	); err != nil {
		log.Fatal(err)
	}

	// 启动消费者后台 Goroutine
	go order.ConsumeOrderMessages()
	go order.ConsumeTimeoutMessages()
	coupon.StartCouponConsumer()
	// **初始化 Gin 服务器**
	r := route.InitServer()

	// **启动服务器**
	route.StartServer(r)
}
