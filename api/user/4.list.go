package user

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
	"strings"
)

type Info struct {
	Code       string `json:"code"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Nickname   string `json:"nickname"`
	Gender     string `json:"gender"`
	RealName   string `json:"real_name"`
	Phone      string `json:"phone"`
	BaseSalary string `json:"base_salary"`
}

// uResponse 结构体（包含用户信息和角色）
type uResponse struct {
	Info
	Roles []map[string]string `json:"roles"` // 用户角色列表
}

// 查询普通用户（非员工）
func listUsers(c *gin.Context) {
	type Req struct {
		Index    int    `form:"index" binding:"required,min=1"`
		Size     int    `form:"size" binding:"required,min=1,max=100"`
		Username string `form:"username"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	// 构造查询
	db := config.DB.Table(model.Users{}).
		Where("is_employee = ?", false).
		Limit(req.Size, (req.Index-1)*req.Size).
		Select("code, username, email, nickname, gender, real_name, phone, base_salary").Desc("created")

	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}

	var users []Info
	count, err := db.FindAndCount(&users)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 返回结果
	response.SuccessWithTotal(c, response.SuccessCode, convertToUserResponse(users), int(count))
}

func listEmployees(c *gin.Context) {
	type Req struct {
		Index    int    `form:"index" binding:"required,min=1"`
		Size     int    `form:"size" binding:"required,min=1,max=100"`
		Username string `form:"username"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	db := config.DB.Table(model.Users{}).
		Where("is_employee = ?", true).
		Limit(req.Size, (req.Index-1)*req.Size).
		Select("code, username, email, nickname, gender, real_name, phone, base_salary").Desc("created")

	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}

	var users []Info
	count, err := db.FindAndCount(&users)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.SuccessWithTotal(c, response.SuccessCode, convertToUserResponse(users), int(count))
}

func convertToUserResponse(users []Info) []uResponse {
	userCodes := extractUserCodes(users)

	// 查询用户角色
	var userRoles []struct {
		UserCode string
		RoleCode string
		RoleName string
	}
	if len(userCodes) > 0 {
		placeholders := strings.Repeat("?,", len(userCodes))
		placeholders = placeholders[:len(placeholders)-1]
		args := interfaceSlice(userCodes)
		_ = config.DB.Table(model.UserRole{}).Alias("ur").
			Join("INNER", []interface{}{model.Role{}, "r"}, "ur.role_code = r.code").
			Select("ur.user_code, r.code as role_code, r.name as role_name").
			Where(fmt.Sprintf("ur.user_code IN (%s)", placeholders), args...).
			Find(&userRoles)
	}

	roleMap := make(map[string][]map[string]string)
	for _, ur := range userRoles {
		roleMap[ur.UserCode] = append(roleMap[ur.UserCode], map[string]string{
			"code": ur.RoleCode, "name": ur.RoleName,
		})
	}

	var responseUsers []uResponse
	for _, user := range users {
		responseUsers = append(responseUsers, uResponse{
			Info:  user,
			Roles: roleMap[user.Code],
		})
	}
	return responseUsers
}

// 获取当前用户的角色
func getUserRole(c *gin.Context) {
	userCode := utils.GetUser(c)
	if userCode == "" {
		response.Success(c, response.Unauthorized, errors.New("未登录"))
		return
	}
	// **查询用户的角色**
	var roles []string
	err := config.DB.Table(model.UserRole{}).Alias("ur").
		Join("INNER", []interface{}{model.Role{}, "r"}, "ur.role_code = r.code").
		Where("ur.user_code = ?", userCode).
		Select("r.code").
		Find(&roles)

	if err != nil {
		response.Success(c, response.ServerError, errors.New("角色查询失败"))
		return
	}
	// **查询用户信息**
	var user model.Users
	exist, err := config.DB.Table(model.Users{}).Where("code =?", userCode).Get(&user)
	if !exist || err != nil {
		response.Success(c, response.QueryFail, fmt.Errorf("查询用户失败 %v", err))
	}
	// **返回角色信息**
	response.SuccessWithData(c, response.SuccessCode, gin.H{
		"user":  user,
		"roles": roles,
	})
}

// extractUserCodes 提取用户 ID 列表
func extractUserCodes(users []Info) []string {
	var codes []string
	for _, user := range users {
		codes = append(codes, user.Code)
	}
	return codes
}

// interfaceSlice 转换 []string 为 []interface{}，用于 SQL 查询
func interfaceSlice(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}
