package salary

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"time"
)

// 处理工资发放
func payUserSalary(c *gin.Context) {
	type Req struct {
		UserCode  string  `json:"user_code" binding:"required"`
		Bonus     float64 `json:"bonus" binding:"required"`     // 奖金
		Deduction float64 `json:"deduction" binding:"required"` // 扣款
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	// 1. 查询用户基本工资
	var user model.Users
	exists, err := config.DB.Where("code = ?", req.UserCode).Get(&user)
	if err != nil || !exists {
		response.Success(c, response.BadRequest, fmt.Errorf("用户不存在"))
		return
	}

	// 2. 计算总工资
	totalSalary := user.BaseSalary + req.Bonus - req.Deduction

	// 3. 记录工资发放
	salaryRecord := model.SalaryRecord{
		UserCode:    req.UserCode,
		BaseSalary:  user.BaseSalary,
		Bonus:       req.Bonus,
		Deduction:   req.Deduction,
		TotalSalary: totalSalary,
		PayDate:     time.Now().Unix(),
	}

	if _, err := config.DB.Insert(&salaryRecord); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
