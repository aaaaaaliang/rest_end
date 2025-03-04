package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// 更新用户信息
func updateUser(c *gin.Context) {
	type Req struct {
		Code       string  `json:"code" binding:"required"`
		Email      string  `json:"email"`
		Nickname   string  `json:"nickname"`
		Gender     string  `json:"gender"`
		RealName   string  `json:"real_name"`
		Phone      string  `json:"phone"`
		Password   string  `json:"password"`
		BaseSalary float64 `json:"base_salary"`
		IsEmployee bool    `json:"is_employee"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}
	if req.Code == "admin" {
		response.Success(c, response.UpdateFail, errors.New("超级管理员不允许修改"))
		return
	}

	//userCode := utils.GetUser(c)
	//if req.Code != userCode {
	//	response.Success(c, response.UpdateFail, errors.New("身份信息不一致"))
	//	return
	//}

	user := model.Users{
		Email:      req.Email,
		Nickname:   req.Nickname,
		Gender:     req.Gender,
		RealName:   req.RealName,
		Phone:      req.Phone,
		BaseSalary: req.BaseSalary,
		IsEmployee: req.IsEmployee,
	}
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			response.Success(c, response.ServerError, err)
			return
		}
		user.Password = hashedPassword
	}

	affected, err := config.DB.Where("code = ?", req.Code).Update(&user)
	if err != nil || affected == 0 {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
