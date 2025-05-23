package cart

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// DeleteCart 删除购物车项
func deleteCart(c *gin.Context) {
	type Req struct {
		Code []string `json:"code" binding:"required"`
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)

	// 删除购物车项
	affectRow, err := config.DB.Where("user_code = ?", userCode).In("code", req.Code).Delete(&model.UserCart{})
	if err != nil {
		response.Success(c, response.DeleteFail, err)
		return
	}
	if affectRow == 0 {
		response.Success(c, response.NotFound, errors.New("购物车项不存在"))
		return
	}

	response.Success(c, response.SuccessCode)
}
