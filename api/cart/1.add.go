package cart

import (
	"fmt"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"

	"github.com/gin-gonic/gin"
)

// addCart 简化后的购物车逻辑
func addCart(c *gin.Context) {
	type Req struct {
		ProductCode string `json:"product_code" binding:"required"`
		ProductNum  int    `json:"product_num" binding:"required,min=1"`
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode, ok := utils.GetUserCode(c)
	if !ok {
		return
	}

	// 1. 获取商品信息（只查数据库）
	product := model.Products{}
	has, err := config.DB.Where("code = ?", req.ProductCode).Get(&product)
	if err != nil || !has {
		response.Success(c, response.QueryFail, fmt.Errorf("商品不存在: %v", req.ProductCode))
		return
	}

	// 2. 判断库存是否足够（此时不扣减库存）
	if product.Count < int64(req.ProductNum) {
		response.Success(c, response.ServerError, fmt.Errorf("库存不足，当前库存: %d", product.Count))
		return
	}

	// 3. 获取用户购物车中该商品的当前记录
	existingCart := model.UserCart{}
	hasCart, err := config.DB.Where("user_code = ? AND product_code = ? AND is_ordered = ?", userCode, req.ProductCode, 0).Get(&existingCart)
	if err != nil {
		response.Success(c, response.ServerError, fmt.Errorf("查询购物车失败: %v", err))
		return
	}

	// 4. 开启事务保存
	session := config.DB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	if hasCart {
		// 已有记录，更新数量和价格
		existingCart.ProductNum = req.ProductNum
		existingCart.TotalPrice = float64(req.ProductNum) * product.Price
		if _, err := session.Where("id = ?", existingCart.Id).Cols("product_num", "total_price").Update(&existingCart); err != nil {
			session.Rollback()
			response.Success(c, response.ServerError, err)
			return
		}
	} else {
		// 插入新记录
		newCart := model.UserCart{
			UserCode:    userCode,
			ProductCode: req.ProductCode,
			ProductNum:  req.ProductNum,
			TotalPrice:  float64(req.ProductNum) * product.Price,
		}
		if _, err := session.Insert(&newCart); err != nil {
			session.Rollback()
			response.Success(c, response.ServerError, err)
			return
		}
	}

	if err := session.Commit(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	log.Printf("✅ 商品 %s 添加购物车成功", req.ProductCode)
	response.Success(c, response.SuccessCode)
}
