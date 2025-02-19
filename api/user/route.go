package user

import "github.com/gin-gonic/gin"

func RegisterUserRoutes(group *gin.RouterGroup) {
	// 创建用户
	group.POST("/user", createUser)

	// 生成验证码
	group.GET("/user/captcha", generateCaptcha)

	// 用户登录
	group.POST("/user/login", login)

	// GitHub OAuth 登录
	group.GET("/user/oauth/github/login", githubLogin)

	// GitHub OAuth 回调
	group.GET("/user/oauth/callback", githubCallback)

	// 新增用户
	group.POST("/user/add", createUsers)

	// 删除用户
	group.DELETE("/user", deleteUser)

	// 更新用户信息
	group.PUT("/user", updateUser)

	// 重置用户密码
	group.PUT("/user/reset", resetUserPassword)

	// 查询用户列表
	group.GET("/user/list", listUsers)

	// 给用户分配角色
	group.POST("/user/assign", assignUserRoles)

	// 查询用户的角色
	group.GET("/user/roles", getUserRoles)

	// 查询用户的所有权限
	group.GET("/user/permissions", getUserPermissions)

	// 查询用户角色信息
	group.GET("/user/role", getUserRole)
	//// 查询用户信息
	//group.GET("/user/info", gerUserInfo)
}
