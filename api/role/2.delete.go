package role

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

func deleteRole(c *gin.Context) {
	type Req struct {
		Code string `json:"code" binding:"required" form:"code"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	affected, err := config.DB.Where("code = ?", req.Code).Delete(&model.Role{})
	if err != nil || affected == 0 {
		response.Success(c, response.DeleteFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}

func removeRolePermission(c *gin.Context) {
	type Req struct {
		RoleCode       string `json:"role_code" binding:"required"`
		PermissionCode string `json:"permission_code" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	affected, err := config.DB.Where("role_code = ? AND permission_code = ?", req.RoleCode, req.PermissionCode).
		Delete(&model.RolePermission{})
	if err != nil || affected == 0 {
		response.Success(c, response.DeleteFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
