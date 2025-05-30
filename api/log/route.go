package log

import "github.com/gin-gonic/gin"

func RegisterLogRoutes(r *gin.RouterGroup) {
	r.GET("/log/es", searchLog)
}
