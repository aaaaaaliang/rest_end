package middleware

import (
	"github.com/gin-gonic/gin"
	"rest/utils"
	"time"
)

func GinLogger(log utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		traceID := utils.NewTraceID()
		ctx := utils.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("trace_id", traceID)

		c.Next()

		latency := time.Since(start)
		log.WithFields(map[string]interface{}{
			"trace_id": traceID,
			"status":   c.Writer.Status(),
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"latency":  latency.String(),
		}).Info("请求日志")
	}
}
