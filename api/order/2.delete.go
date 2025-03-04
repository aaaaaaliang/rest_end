package order

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func deleteOrder(c *gin.Context) {
	type Req struct {
		Code string `form:"code" json:"code" binding:"required"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	fmt.Println("code", req.Code)

	if affectRow, err := config.DB.Where("code = ?", req.Code).Delete(model.UserOrder{}); err != nil || affectRow != 1 {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
