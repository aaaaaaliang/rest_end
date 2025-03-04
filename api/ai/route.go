package ai

import "github.com/gin-gonic/gin"

func RegisterAIRoutes(group *gin.RouterGroup) {
	group.POST("/ai", chatWithAI)
}
