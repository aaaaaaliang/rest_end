package dashboard

import "github.com/gin-gonic/gin"

func RegisterDashboardRoutes(group *gin.RouterGroup) {
	group.GET("/dashboard", countTodayOrders)
}
