package product

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// ListProducts 查询产品列表（支持分页 & 分类过滤）
func listProducts(c *gin.Context) {
	type Req struct {
		Index        int    `form:"index" binding:"required,min=1"` // 页码
		Size         int    `form:"size" binding:"required,min=1"`  // 每页大小
		CategoryCode string `form:"category_code" binding:"omitempty"`
		Main         int    `form:"main"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	var products []struct {
		model.Products `xorm:"extends"`
		CategoryName   string `json:"category_name"` // 分类名称
	}

	// 初始化查询
	db := config.DB.Table("products").
		Join("LEFT", "category", "products.category_code = category.code").
		Select("products.*, category.category_name").
		Limit(req.Size, (req.Index-1)*req.Size).
		Desc("products.created")

	// 过滤分类
	if req.CategoryCode != "" {
		db = db.Where("products.category_code = ?", req.CategoryCode)
	}

	// 只有当 `Main == 1` 时，才添加 `WHERE main = 1`
	if req.Main == 1 {
		db = db.Where("products.main = ?", 1)
	}

	// 执行查询
	total, err := db.FindAndCount(&products)
	if err != nil {
		response.Success(c, response.QueryFail, err)
		return
	}

	response.SuccessWithTotal(c, response.SuccessCode, products, int(total))
}
