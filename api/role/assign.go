package role

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func assignRolePermissions(c *gin.Context) {
	type Req struct {
		RoleCode        string   `json:"role_code" binding:"required"`
		PermissionCodes []string `json:"permission_codes" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	if req.RoleCode == "58732ecc-7ed5-49a7-8603-f721be698e90" {
		userCode, exist := utils.GetUserCode(c)
		if !exist {
			return
		}

		if userCode != "admin" {
			response.Success(c, response.UpdateFail, errors.New("只有管理员才能操作admin"))
			return
		}
	}

	// 删除已有权限
	_, err := config.DB.Where("role_code = ?", req.RoleCode).Delete(&model.RolePermission{})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 插入新的权限
	for _, permCode := range req.PermissionCodes {
		rolePerm := model.RolePermission{
			RoleCode:       req.RoleCode,
			PermissionCode: permCode,
		}
		if _, err := config.DB.Insert(&rolePerm); err != nil {
			response.Success(c, response.ServerError, err)
			return
		}
	}

	response.Success(c, response.SuccessCode)
}
