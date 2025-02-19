package role

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

func updateRole(c *gin.Context) {
	type Req struct {
		Code        string `json:"code" binding:"required"`
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

	affected, err := config.DB.Where("code = ?", req.Code).Update(&role)
	if err != nil || affected == 0 {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
