package banner

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
)

// 删除轮播图
func deleteBanner(c *gin.Context) {
	type Req struct {
		Code string `json:"code" binding:"required" form:"code"`
	}

	var req Req
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Success(c, response.BadRequest, err)
		return
	}

	if affectRow, err := config.DB.Where("code = ?", req.Code).Delete(&model.Banner{}); err != nil || affectRow != 1 {
		response.Success(c, response.ServerError, fmt.Errorf("%v 删除失败", err))
		return
	}

	response.Success(c, response.SuccessCode)
}
