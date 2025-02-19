package user

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func createUser(c *gin.Context) {
	type Req struct {
		Username        string `json:"username" binding:"required"`
		Password        string `json:"password" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		CaptchaID       string `json:"captcha_id" binding:"required"`
		CaptchaSolution string `json:"captcha" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	// 使用 base64Captcha 默认存储进行验证码验证
	if !base64Captcha.DefaultMemStore.Verify(req.CaptchaID, req.CaptchaSolution, true) {
		response.Success(c, response.BadRequest, errors.New("验证码错误"))
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	user := model.Users{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}

	// 检查用户名是否已存在
	exist, err := config.DB.Table(model.Users{}).Where("username = ?", req.Username).Exist()
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	if exist {
		response.Success(c, response.BadRequest, errors.New("用户已存在"))
		return
	}

	// 插入用户
	if _, err = config.DB.Table(model.Users{}).Insert(&user); err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("创建用户失败: %v", err))
		return
	}

	response.Success(c, response.SuccessCode)
}

// 新增用户 主要是员工
func createUsers(c *gin.Context) {
	type Req struct {
		Username   string  `json:"username" binding:"required"`
		Password   string  `json:"password"` // 如果不传默认就是 111111
		Email      string  `json:"email"`
		Nickname   string  `json:"nickname"`
		Gender     string  `json:"gender"`
		RealName   string  `json:"real_name"`
		Phone      string  `json:"phone"`
		BaseSalary float64 `json:"base_salary"`
		IsEmployee bool    `json:"is_employee"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}
	if req.Password == "" {
		req.Password = "111111"
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	exist, err := config.DB.Table(model.Users{}).Where("username = ?", req.Username).Exist()
	if err != nil {
		response.Success(c, response.QueryFail, err)
		return
	}
	if exist {
		response.Success(c, response.UserExists)
	}
	user := model.Users{
		Username:   req.Username,
		Password:   hashedPassword,
		Email:      req.Email,
		Nickname:   req.Nickname,
		Gender:     req.Gender,
		RealName:   req.RealName,
		Phone:      req.Phone,
		BaseSalary: req.BaseSalary,
		IsEmployee: req.IsEmployee,
	}
	log.Println("user", user)

	if _, err := config.DB.Insert(&user); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}

// 查询用户的角色，返回所有角色，并标记当前用户拥有的角色
func getUserRoles(c *gin.Context) {
	type Req struct {
		UserCode string `form:"user_code" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	// 查询所有角色
	var allRoles []model.Role
	if err := config.DB.Table(model.Role{}).Find(&allRoles); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 查询用户已有的角色
	var userRoleCodes []string
	err := config.DB.Table(model.UserRole{}).Alias("ur").
		Where("ur.user_code = ?", req.UserCode).
		Select("ur.role_code").
		Find(&userRoleCodes)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 构造 userRoleCodes 的哈希表，加快查找速度
	roleMap := make(map[string]struct{}, len(userRoleCodes))
	for _, code := range userRoleCodes {
		roleMap[code] = struct{}{}
	}

	type RoleResponse struct {
		Code    string `json:"code"`
		Name    string `json:"name"`
		Checked bool   `json:"checked"` // 是否勾选
	}

	var responseRoles []RoleResponse
	for _, role := range allRoles {
		_, exists := roleMap[role.Code] // 检查该角色是否存在于用户拥有的角色列表
		responseRoles = append(responseRoles, RoleResponse{
			Code:    role.Code,
			Name:    role.Name,
			Checked: exists, // 如果存在，则 checked 为 true
		})
	}

	response.SuccessWithData(c, response.SuccessCode, responseRoles)
}

// 查询用户的所有权限（合并多个角色的权限）
func getUserPermissions(c *gin.Context) {
	type Req struct {
		UserCode string `form:"user_code" binding:"required"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	var permissions []model.APIPermission
	err := config.DB.Table("user_role").Alias("ur").
		Join("INNER", []interface{}{model.RolePermission{}, "rp"}, "ur.role_code = rp.role_code").
		Join("INNER", []interface{}{model.APIPermission{}, "p"}, "rp.permission_code = p.code").
		Where("ur.user_code = ?", req.UserCode).
		Distinct("p.*").
		Find(&permissions)

	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.SuccessWithData(c, response.SuccessCode, permissions)
}

// 重置用户密码
func resetUserPassword(c *gin.Context) {
	type Req struct {
		UserCode string `json:"user_code" form:"user_code"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}
	// 加密密码
	hashedPassword, err := utils.HashPassword("111111")
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	// 将所有用户的密码字段置为空字符串
	_, err = config.DB.Table(model.Users{}).Where("code = ?", req.UserCode).Update(map[string]interface{}{
		"password": hashedPassword,
	})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
