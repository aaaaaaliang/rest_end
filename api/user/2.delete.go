package user

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 删除用户
func deleteUser(c *gin.Context) {
	type Req struct {
		UserCode string `json:"user_code" binding:"required" form:"user_code"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	// 事务保证一致性
	session := config.DB.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 删除用户
	if _, err := session.Where("code = ?", req.UserCode).Delete(&model.Users{}); err != nil {
		session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// 删除用户的角色关联
	if _, err := session.Where("user_code = ?", req.UserCode).Delete(&model.UserRole{}); err != nil {
		session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	session.Commit()
	response.Success(c, response.SuccessCode)
}
