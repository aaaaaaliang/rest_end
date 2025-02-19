package product

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// UpdateProduct 更新产品
func updateProduct(c *gin.Context) {
	type Req struct {
		Code         string      `json:"code" binding:"required"`
		ProductsName string      `json:"products_name" binding:"required"`
		CategoryCode string      `json:"category_code" binding:"required"`
		Price        float64     `json:"price" binding:"required"`
		Count        int64       `json:"count" binding:"required"`
		Describe     string      `json:"describe"`
		Picture      model.Annex `json:"picture"`
		Main         int         `json:"main"`
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	// 检查产品是否存在
	exist, err := config.DB.Where("code = ?", req.Code).Exist(&model.Products{})
	if err != nil {
		response.Success(c, response.ServerError, errors.New("查询产品失败"))
		return
	}

	if !exist {
		response.Success(c, response.NotFound, errors.New("产品不存在"))
		return
	}

	// 执行更新
	affectRow, err := config.DB.Where("code = ?", req.Code).Update(&model.Products{
		ProductsName: req.ProductsName,
		CategoryCode: req.CategoryCode,
		Price:        req.Price,
		Count:        req.Count,
		Describe:     req.Describe,
		Picture:      req.Picture,
		Main:         req.Main,
	})
	if err != nil {
		response.Success(c, response.UpdateFail, err)
		return
	}

	if affectRow == 0 {
		response.Success(c, response.UpdateFail, errors.New("产品未更新"))
		return
	}

	response.Success(c, response.SuccessCode)
}
