package order

import "github.com/gin-gonic/gin"

func RegisterOrderRoutes(group *gin.RouterGroup) {
	group.POST("/order", addOrder)
	group.DELETE("/order", deleteOrder)
	group.PUT("/order", updateOrder)
	group.GET("/order", listOrder)
}
