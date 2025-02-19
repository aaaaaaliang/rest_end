package middleware

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"strings"
)

// Cors 处理跨域
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取 CORS 配置
		allowedOrigins := config.G.Cors.AllowOrigins
		allowCredentials := config.G.Cors.AllowCredentials

		origin := c.GetHeader("Origin") // 获取请求来源
		if origin != "" {
			// 只有在配置的域名列表中，才允许访问
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == origin {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// 允许的方法
		c.Header("Access-Control-Allow-Methods", strings.Join(config.G.Cors.AllowMethods, ","))
		// 允许的头部
		c.Header("Access-Control-Allow-Headers", strings.Join(config.G.Cors.AllowHeaders, ","))

		// 是否允许携带 Cookie
		if allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 预检请求处理
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // 204 No Content
			return
		}

		c.Next()
	}
}
