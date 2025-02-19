package public

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes **在这里定义 Public 相关的路由**
func RegisterPublicRoutes(group *gin.RouterGroup) {
	group.POST("/upload", uploadFile)
}
