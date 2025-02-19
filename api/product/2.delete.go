package product

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// DeleteProduct 删除产品
func deleteProduct(c *gin.Context) {
	type Req struct {
		Code string `json:"code" binding:"required" form:"code"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	affectRow, err := config.DB.Where("code = ?", req.Code).Delete(&model.Products{})
	if err != nil {
		response.Success(c, response.DeleteFail, err)
		return
	}

	if affectRow == 0 {
		response.Success(c, response.NotFound, errors.New("产品不存在"))
		return
	}

	response.Success(c, response.SuccessCode)
}
