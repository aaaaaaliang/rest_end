package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/response"
	"strings"
)

// ValidationQuery 校验 URL 查询参数
func ValidationQuery(c *gin.Context, d any) (success bool) {
	err := c.ShouldBindQuery(d)
	if err != nil && strings.Contains(err.Error(), "required") {
		response.Success(c, response.NotFound, err)
		success = false
		return
	}
	if err != nil && strings.Contains(err.Error(), "validation") {
		response.Resp(c, response.BadRequest, err)
		success = false
		return
	}
	if err != nil {
		response.Resp(c, response.BadRequest, err)
		success = false
		return
	}
	success = true
	return
}

// ValidationJson 校验 JSON 数据
func ValidationJson(c *gin.Context, d any) (success bool) {
	err := c.ShouldBindJSON(d)
	fmt.Println("1") // 调试信息
	if err != nil && strings.Contains(err.Error(), "required") {
		response.Resp(c, response.NotFound, err)
		success = false
		return
	}
	fmt.Println("2") // 调试信息
	if err != nil && strings.Contains(err.Error(), "validation") {
		response.Resp(c, response.BadRequest, err)
		success = false
		return
	}
	fmt.Println("3") // 调试信息
	if err != nil {
		fmt.Println("err:", err)
		response.Resp(c, response.BadRequest, err)
		success = false
		return
	}
	success = true
	return
}
