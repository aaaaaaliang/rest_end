package ai_v1_user

import (
	"github.com/gin-gonic/gin"
)

func RegisterAIModelRoutes(group *gin.RouterGroup) {
	group.POST("/ai-model", addAIModel)
	group.GET("/ai-model", getAIModel)
	group.PUT("/ai-model", updateAIModel)
	//group.POST("/api/ai/chat", chatWithAI)
}
