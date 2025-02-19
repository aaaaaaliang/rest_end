package utils

import "github.com/gin-gonic/gin"

func GetUser(c *gin.Context) string {
	return c.GetString("user")
}
