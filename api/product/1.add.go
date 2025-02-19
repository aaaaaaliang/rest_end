package product

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// AddProduct 添加产品
func addProduct(c *gin.Context) {
	type Req struct {
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

	product := model.Products{
		ProductsName: req.ProductsName,
		CategoryCode: req.CategoryCode,
		Price:        req.Price,
		Count:        req.Count,
		Describe:     req.Describe,
		Picture:      req.Picture,
		Main:         req.Main,
	}
	if _, err := config.DB.Insert(&product); err != nil {
		response.Success(c, response.CreateFail, err)
		return
	}
	response.Success(c, response.SuccessCode)
}
