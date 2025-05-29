package ai_v1_user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// 创建模型配置（只允许插入一条）
func addAIModel(c *gin.Context) {
	var req model.AIModelConfig
	if !utils.ValidationJson(c, &req) {
		return
	}

	// 限制只能有一条模型配置记录
	cnt, _ := config.DB.Count(&model.AIModelConfig{})
	if cnt > 0 {
		response.Success(c, response.BadRequest, errors.New("已有模型配置，请使用更新接口"))
		return
	}

	_, err := config.DB.Insert(&req)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.Success(c, response.SuccessCode)
}
