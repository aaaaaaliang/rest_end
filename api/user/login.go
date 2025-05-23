package user

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func login(c *gin.Context) {
	type Req struct {
		Username        string `json:"username" binding:"required"`
		Password        string `json:"password" binding:"required"`
		CaptchaID       string `json:"captcha_id" binding:"required"`
		CaptchaSolution string `json:"captcha" binding:"required"`
	}

	var req Req
	if ok := utils.ValidationJson(c, &req); !ok {
		return
	}
	// 使用 base64Captcha 默认存储进行验证码验证
	if !base64Captcha.DefaultMemStore.Verify(req.CaptchaID, req.CaptchaSolution, true) {
		response.Success(c, response.BadRequest, errors.New("验证码错误"))
		return
	}

	// 查询用户
	var user model.Users
	has, err := config.DB.Where("username = ?", req.Username).Get(&user)
	if err != nil || !has {
		response.Success(c, response.ServerError, fmt.Errorf("用户名或密码错误或者发生%v", err))
		return
	}

	// 校验密码
	if !utils.CheckPassword(req.Password, user.Password) {
		response.Success(c, response.Unauthorized, errors.New("用户名或密码错误"))
		return
	}

	// 生成 JWT Token
	token, err := config.GenerateJWT(user.Code)
	if err != nil {
		response.Success(c, response.ServerError, errors.New("生成 Token 失败"))
		return
	}

	// 设置 Cookie
	c.SetCookie("access_token", token, 3600, "/", "", false, true)

	// 登录成功
	response.Success(c, response.SuccessCode)
}

func logout(c *gin.Context) {
	// 删除 Cookie
	c.SetCookie("access_token", "", -1, "/", "", false, true) // 设置过期时间为过去，清除 Cookie

	// 退出成功
	response.Success(c, response.SuccessCode)
}
