package utils

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
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
		response.Success(c, response.BadRequest, err)
		success = false
		return
	}
	if err != nil {
		response.Success(c, response.BadRequest, err)
		success = false
		return
	}
	success = true
	return
}

// ValidationJson 校验 JSON 数据
func ValidationJson(c *gin.Context, d any) (success bool) {
	log.Println("Request Body:", c.Request.Body) // 打印请求体内容

	// 先检查请求体是否为空
	if c.Request.Body == nil {
		response.Success(c, response.BadRequest, errors.New("请求体为空"))
		success = false
		return
	}

	err := c.ShouldBindJSON(d)
	fmt.Println("1") // 调试信息
	if err != nil && strings.Contains(err.Error(), "required") {
		response.Success(c, response.NotFound, err)
		success = false
		return
	}
	fmt.Println("2") // 调试信息
	if err != nil && strings.Contains(err.Error(), "validation") {
		response.Success(c, response.BadRequest, err)
		success = false
		return
	}
	fmt.Println("3") // 调试信息
	if err != nil {
		fmt.Println("err:", err)
		response.Success(c, response.BadRequest, err)
		success = false
		return
	}

	success = true
	return
}
