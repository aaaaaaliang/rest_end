package user

import (
	"fmt"
	"rest/response"

	"github.com/gin-gonic/gin"
)

// GetUsers 获取所有用户
func getUsers(c *gin.Context) {
	fmt.Println("GetUsers 方法被调用")
	c.JSON(200, gin.H{"message": "User list"})
}

// GetUserByID 获取单个用户
func getUserByID(c *gin.Context) {
	id := c.Param("id")
	fmt.Println("GetUserByID 方法被调用，ID:", id)
	c.JSON(200, gin.H{"message": "User details", "id": id})
}

// CreateUser 创建用户
func createUser(c *gin.Context) {
	fmt.Println("CreateUser 方法被调用")
	c.JSON(200, gin.H{"message": "User created"})
}

// UpdateUser 更新用户
func updateUser(c *gin.Context) {
	id := c.Param("id")
	fmt.Println("UpdateUser 方法被调用，ID:", id)
	c.JSON(200, gin.H{"message": "User updated", "id": id})
}

// DeleteUser 删除用户
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	fmt.Println("DeleteUser 方法被调用，ID:", id)
	c.JSON(200, gin.H{"message": "User deleted", "id": id})
}


// CreateUser 创建新用户，并分配角色
func CreateUser(c *gin.Context) {
	type Req struct {
		Username        string `json:"username" binding:"required"`
		Password        string `json:"password" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		CaptchaID       string `json:"captcha_id"`
		CaptchaSolution string `json:"captcha_solution"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		response.SuccessWithData(c, response.BadRequestCode, err.Error())
		return
	}

	// 验证验证码
	if !base64Captcha.DefaultMemStore.Verify(req.CaptchaID, req.CaptchaSolution, false) {
		public.Resp(c, public.StatusInvalidParams, errors.New("验证码错误"))
		return
	}

	// 加密密码
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		public.Resp(c, public.StatusInternalServerError, err)
		return
	}

	user := model.Users{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}

	// 检查用户名是否存在
	exist, err := public.DB.Table(model.Users{}).Where("username = ?", req.Username).Exist()
	if err != nil || exist {
		public.Resp(c, public.StatusDataUpdateError, errors.New("查询用户失败 或用户已存在"))
		return
	}

	// 插入用户
	if _, err = public.DB.Insert(&user); err != nil {
		public.Resp(c, public.StatusDataInsertError, fmt.Errorf("创建用户失败: %v", err))
		return
	}

	public.Resp(c, public.StatusOK)
}

