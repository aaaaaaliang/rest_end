package ai_v1_user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

func getAIModel(c *gin.Context) {
	var m model.AIModelConfig
	has, err := config.DB.Get(&m)
	if err != nil || !has {
		response.Success(c, response.ServerError, errors.New("未找到模型配置"))
		return
	}
	response.SuccessWithData(c, response.SuccessCode, m)
}
