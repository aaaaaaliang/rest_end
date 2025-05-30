package cart

import (
	"fmt"
	"rest/config"
	logger "rest/logger"
	"rest/model"
	"rest/response"
	"rest/utils"

	"github.com/gin-gonic/gin"
)

// 添加购物车
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
		logger.SendLogToESCtx(c.Request.Context(), "WARN", "cart", "error", "cart.add.no_user", nil)
		return
	}

	// 获取商品信息
	product := model.Products{}
	has, err := config.DB.Where("code = ?", req.ProductCode).Get(&product)
	if err != nil || !has {
		logger.SendLogToESCtx(c.Request.Context(), "WARN", "cart", "error", "cart.add.product_not_exist", map[string]interface{}{
			"product_code": req.ProductCode,
			"err":          err,
		})
		response.Success(c, response.QueryFail, fmt.Errorf("商品不存在: %v", req.ProductCode))
		return
	}

	// 库存校验
	if product.Count < int64(req.ProductNum) {
		logger.SendLogToESCtx(c.Request.Context(), "WARN", "cart", "error", "cart.add.stock_not_enough", map[string]interface{}{
			"product_code": req.ProductCode,
			"stock":        product.Count,
			"want":         req.ProductNum,
		})
		response.Success(c, response.ServerError, fmt.Errorf("库存不足，当前库存: %d", product.Count))
		return
	}

	// 查询是否已有购物车记录
	existingCart := model.UserCart{}
	hasCart, err := config.DB.Where("user_code = ? AND product_code = ? AND is_ordered = ?", userCode, req.ProductCode, 0).Get(&existingCart)
	if err != nil {
		logger.SendLogToESCtx(c.Request.Context(), "ERROR", "cart", "error", "cart.add.query_cart_fail", map[string]interface{}{
			"user_code":    userCode,
			"product_code": req.ProductCode,
			"err":          err,
		})
		response.Success(c, response.ServerError, fmt.Errorf("查询购物车失败: %v", err))
		return
	}

	// 开启事务保存
	session := config.DB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		logger.SendLogToESCtx(c.Request.Context(), "ERROR", "cart", "error", "cart.add.tx_begin_fail", map[string]interface{}{
			"err": err,
		})
		response.Success(c, response.ServerError, err)
		return
	}

	if hasCart {
		existingCart.ProductNum = req.ProductNum
		existingCart.TotalPrice = float64(req.ProductNum) * product.Price
		if _, err := session.Where("id = ?", existingCart.Id).Cols("product_num", "total_price").Update(&existingCart); err != nil {
			session.Rollback()
			logger.SendLogToESCtx(c.Request.Context(), "ERROR", "cart", "error", "cart.add.update_cart_fail", map[string]interface{}{
				"err": err,
			})
			response.Success(c, response.ServerError, err)
			return
		}
	} else {
		newCart := model.UserCart{
			UserCode:    userCode,
			ProductCode: req.ProductCode,
			ProductNum:  req.ProductNum,
			TotalPrice:  float64(req.ProductNum) * product.Price,
		}
		if _, err := session.Insert(&newCart); err != nil {
			session.Rollback()
			logger.SendLogToESCtx(c.Request.Context(), "ERROR", "cart", "error", "cart.add.insert_cart_fail", map[string]interface{}{
				"err": err,
			})
			response.Success(c, response.ServerError, err)
			return
		}
	}

	if err := session.Commit(); err != nil {
		logger.SendLogToESCtx(c.Request.Context(), "ERROR", "cart", "error", "cart.add.tx_commit_fail", map[string]interface{}{
			"err": err,
		})
		response.Success(c, response.ServerError, err)
		return
	}

	// ✅ 成功日志
	logger.SendLogToESCtx(c.Request.Context(), "INFO", "cart", "operation", "cart.add.success", map[string]interface{}{
		"product_code":  product.Code,
		"product_name":  product.ProductsName,
		"product_price": product.Price,
		"product_num":   req.ProductNum,
		"product_total": float64(req.ProductNum) * product.Price,
		"stock_remain":  product.Count,
	})

	response.Success(c, response.SuccessCode)
}
