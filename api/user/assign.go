package user

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// 给用户分配角色
func assignUserRoles(c *gin.Context) {
	type Req struct {
		UserCode  string   `json:"user_code" binding:"required"`
		RoleCodes []string `json:"role_codes" binding:"required"`
	}

	var req Req
	if ok := utils.ValidationJson(c, &req); !ok {
		return
	}

	session := config.DB.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 删除已有的角色
	if _, err := session.Where("user_code = ?", req.UserCode).Delete(&model.UserRole{}); err != nil {
		session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// 插入新角色
	for _, roleCode := range req.RoleCodes {
		userRole := model.UserRole{
			UserCode: req.UserCode,
			RoleCode: roleCode,
		}
		if _, err := session.Insert(&userRole); err != nil {
			session.Rollback()
			response.Success(c, response.ServerError, err)
			return
		}
	}

	session.Commit()
	response.Success(c, response.SuccessCode)
}
