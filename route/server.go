package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"rest/api/ws"
	"rest/config"
	"rest/middleware"
)

// InitServer 初始化 Gin 服务器
func InitServer() *gin.Engine {
	// **创建 Gin 实例**
	r := gin.Default()

	// **加载中间件**
	loadMiddlewares(r)

	// **注册 API 路由**
	registerRoutes(r)
	// ** ws **
	// **创建 WebSocket Hub**
	//ctx, cancel := context.WithCancel(context.Background())
	//ctx := context.Background()
	hub := ws.NewHub()
	//// **监听系统信号，优雅退出**
	//go func() {
	//	sig := make(chan os.Signal, 1)
	//	signal.Notify(sig, os.Interrupt)
	//	<-sig
	//	log.Println("服务器关闭，停止 Redis 监听")
	//	cancel() // **取消 context，确保 Redis 订阅自动退出**
	//}()
	// **手动注册 WebSocket 路由**

	r.GET("/api/ws", func(c *gin.Context) {
		log.Println("Received WebSocket connection request")
		ws.ServeWs(hub, c)
	})
	// **自动注册 API 权限**
	autoRegisterAPIPermissions(r)

	return r
}

// loadMiddlewares 统一加载所有中间件
func loadMiddlewares(r *gin.Engine) {
	r.Use(middleware.Cors())                 // 跨域
	r.Use(middleware.PermissionMiddleware()) // 权限控制

	// 静态文件目录（如图片/上传文件）
	r.Static("/uploads", config.G.Uploads.Url)
}

// StartServer 启动服务器
func StartServer(r *gin.Engine) {
	host := config.G.App.Host
	port := config.G.App.Port
	fmt.Printf("🚀 服务启动: %v:%d\n", host, port)
	log.Fatal(r.Run(fmt.Sprintf(":%d", port)))
}
