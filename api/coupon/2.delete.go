package coupon

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

func deleteTemplate(c *gin.Context) {
	code := c.Param("code")
	affect, err := config.DB.Where("code = ?", code).Delete(&model.CouponTemplate{})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	if affect == 0 {
		response.Success(c, response.NotFound, errors.New("模板不存在或已删除"))
		return
	}
	response.Success(c, response.SuccessCode)
}
