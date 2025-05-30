package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/logger"
	"rest/response"
	"strings"
)

// ValidationQuery 校验 URL 查询参数
func ValidationQuery(c *gin.Context, d any) (success bool) {
	err := c.ShouldBindQuery(d)
	if err != nil {
		logValidationError(c, "query", err, d)
		status := response.BadRequest
		if strings.Contains(err.Error(), "required") {
			status = response.NotFound
		}
		response.Success(c, status, err)
		return false
	}
	return true
}

// ValidationJson 校验 JSON 数据
func ValidationJson(c *gin.Context, d any) (success bool) {
	if c.Request.Body == nil {
		msg := "请求体为空"
		logger.SendLogToESCtx(c.Request.Context(), "WARN", "validation", "error", "json.empty_body", nil)
		response.Success(c, response.BadRequest, errors.New(msg))
		return false
	}

	err := c.ShouldBindJSON(d)
	if err != nil {
		logValidationError(c, "json", err, d)
		status := response.BadRequest
		if strings.Contains(err.Error(), "required") {
			status = response.NotFound
		}
		response.Success(c, status, err)
		return false
	}

	return true
}

// logValidationError 写入校验失败日志
func logValidationError(c *gin.Context, source string, err error, req any) {
	logger.SendLogToESCtx(c.Request.Context(), "WARN", "validation", "error", "param.validation_fail", map[string]interface{}{
		"source": source,      // json / query
		"error":  err.Error(), // 具体错误信息
		"req":    req,         // 结构体内容
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	})
}
