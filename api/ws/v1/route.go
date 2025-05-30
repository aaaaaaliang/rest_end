package handler

import "github.com/gin-gonic/gin"

func RegisterWSRoutes(group *gin.RouterGroup) {
	group.GET("chat/customer", createSession)
	group.GET("chat/agent", startSessionByAgent)
	//group.GET("chat", chatWebSocket)
}
