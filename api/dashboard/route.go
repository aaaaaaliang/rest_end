package dashboard

import "github.com/gin-gonic/gin"

// RegisterDashboardRoutes 注册购物车相关路由
func RegisterDashboardRoutes(group *gin.RouterGroup) {
	group.GET("/dashboard", countTodayOrders) // 添加购物车
}
