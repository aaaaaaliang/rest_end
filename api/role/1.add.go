package role

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// 创建角色
func addRole(c *gin.Context) {
	type Req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	var req Req
	if ok := utils.ValidationJson(c, &req); !ok {
		return
	}
	exist, err := config.DB.Table(model.Role{}).Where("name = ?", req.Name).Exist()
	if exist || err != nil {
		response.Success(c, response.QueryFail, fmt.Errorf("角色名称已存在或者 err: %v", err))
		return
	}

	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if _, err := config.DB.Insert(&role); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
