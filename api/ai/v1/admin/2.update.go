package ai_v1_user

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

func updateAIModel(c *gin.Context) {
	var req model.AIModelConfig
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		response.Success(c, response.BadRequest, errors.New("参数错误或缺少 code"))
		return
	}

	// 开启事务
	session := config.DB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("开启事务失败: %v", err))
		return
	}

	// 加行锁（禁止其他并发读取或修改）
	var existing model.AIModelConfig
	has, err := session.Where("code = ?", req.Code).ForUpdate().Get(&existing)
	if err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, fmt.Errorf("查询模型失败: %v", err))
		return
	}
	if !has {
		_ = session.Rollback()
		response.Success(c, response.BadRequest, errors.New("未找到模型"))
		return
	}

	// 进行字段更新（你可以按字段拷贝，也可以用结构体整体覆盖）
	existing.ModelName = req.ModelName
	existing.PromptIntro = req.PromptIntro
	existing.UserLabel = req.UserLabel
	existing.AssistantLabel = req.AssistantLabel
	existing.MaxHistory = req.MaxHistory

	if _, err := session.Where("code = ?", existing.Code).Update(&existing); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, fmt.Errorf("更新失败: %v", err))
		return
	}

	if err := session.Commit(); err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("提交失败: %v", err))
		return
	}

	response.Success(c, response.SuccessCode)
}
