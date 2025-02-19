package role

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 创建角色
func addRole(c *gin.Context) {
	type Req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
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
