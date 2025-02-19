package salary

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 删除薪资记录
func deleteSalaryRecord(c *gin.Context) {
	type Req struct {
		Code string `form:"code" binding:"required"` // 薪资记录ID
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	// 删除记录
	_, err := config.DB.Where("code = ?", req.Code).Delete(&model.SalaryRecord{})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
