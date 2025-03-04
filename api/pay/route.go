package pay

import "github.com/gin-gonic/gin"

func RegisterPayRoutes(group *gin.RouterGroup) {
	group.GET("/pay", payHandler)
	group.POST("/pay/callback", payNotifyHandler)
}
