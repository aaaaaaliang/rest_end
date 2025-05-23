package role

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func updateRole(c *gin.Context) {
	type Req struct {
		Code        string `json:"code" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	var req Req
	if ok := utils.ValidationJson(c, &req); !ok {
		return
	}

	if req.Code == "58732ecc-7ed5-49a7-8603-f721be698e90" {
		userCode, exist := utils.GetUserCode(c)
		if !exist {
			return
		}

		if userCode != "admin" {
			response.Success(c, response.UpdateFail, errors.New("只有管理员才能操作admin"))
			return
		}
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

	affected, err := config.DB.Where("code = ?", req.Code).Update(&role)
	if err != nil || affected == 0 {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
