package role

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func deleteRole(c *gin.Context) {
	type Req struct {
		Code string `form:"code" binding:"required"`
	}

	var req Req
	if ok := utils.ValidationQuery(c, &req); !ok {
		return
	}

	// 1. 检查是否有用户关联该角色
	count, err := config.DB.Where("role_code = ?", req.Code).Count(&model.UserRole{})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	if count > 0 {
		response.Success(c, response.DeleteFail, fmt.Errorf("该角色已分配给 %d 个用户，无法删除", count))
		return
	}

	// 2. 删除角色权限
	if _, err := config.DB.Where("role_code = ?", req.Code).Delete(&model.RolePermission{}); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 3. 删除角色本身
	if _, err := config.DB.Where("code = ?", req.Code).Delete(&model.Role{}); err != nil {
		response.Success(c, response.ServerError, err)
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
	if ok := utils.ValidationJson(c, &req); !ok {
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
