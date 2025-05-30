package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"rest/logger"
	"time"
)

func GinLogger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		traceID := logger.NewTraceID()
		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("trace_id", traceID)

		ctx = context.WithValue(ctx, "gin_context", c)
		c.Request = c.Request.WithContext(ctx)

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
