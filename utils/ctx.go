package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/response"
)

func GetUser(c *gin.Context) string {
	return c.GetString("user")
}

// GetUserCode 获取用户 code 的通用函数
func GetUserCode(c *gin.Context) (string, bool) {
	userCode := c.GetString("user")
	if userCode == "" {
		response.Success(c, response.Unauthorized, errors.New("未拿到用户code"))
		return "", false
	}
	return userCode, true
}
