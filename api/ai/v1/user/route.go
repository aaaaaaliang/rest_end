package ai_v1_user

import "github.com/gin-gonic/gin"

func RegisterAIUserRoutes(group *gin.RouterGroup) {
	// 首次发起会话并聊天
	group.POST("/ai/chat/first", chatFirstMessage)

	// 查询当前用户所有会话列表
	group.GET("/ai/chat/session", listAIChatSessions)

	// 删除某个会话
	group.DELETE("/ai/chat/session", deleteAIChatSession)

	group.PUT("/ai/chat/session", updateAIChatSessionTitle)
	group.POST("/ai/chat/in-session", chatInSession)

	group.GET("/ai/chat/history", getChatHistoryBySession)

}
