package cart

import (
	"errors"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// AddCart 添加购物车
func addCart(c *gin.Context) {
	type Req struct {
		ProductCode string `json:"product_code" binding:"required"`
		ProductNum  int    `json:"product_num" binding:"required"`
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
	if userCode == "" {
		response.Success(c, response.Unauthorized, errors.New("未拿到用户code"))
		return
	}

	// 查询商品信息
	var product model.Products
	exists, err := config.DB.Where("code = ?", req.ProductCode).Get(&product)
	if err != nil {
		response.Success(c, response.QueryFail, err)
		return
	}
	if !exists {
		response.Success(c, response.NotFound, errors.New("商品不存在"))
		return
	}

	// 开启事务
	session := config.DB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 查询购物车项
	existingCart := model.UserCart{
		UserCode:    userCode,
		ProductCode: req.ProductCode,
	}
	exist, err := session.Where("user_code = ? AND product_code = ? AND is_ordered = ?", userCode, req.ProductCode, false).Get(&existingCart)
	if err != nil {
		session.Rollback()
		response.Success(c, response.QueryFail, err)
		return
	}

	if exist {
		// 更新数量
		existingCart.ProductNum = req.ProductNum
		existingCart.TotalPrice = float64(existingCart.ProductNum) * product.Price

		_, err := session.Where("user_code = ? AND product_code = ?", userCode, req.ProductCode).Cols("product_num", "total_price").Update(&existingCart)
		if err != nil {
			session.Rollback()
			response.Success(c, response.UpdateFail, err)
			return
		}
	} else {
		// 创建新的购物车项
		newCart := model.UserCart{
			UserCode:    userCode,
			ProductCode: req.ProductCode,
			ProductNum:  req.ProductNum,
			TotalPrice:  float64(req.ProductNum) * product.Price,
		}
		_, err := session.Insert(&newCart)
		if err != nil {
			session.Rollback()
			response.Success(c, response.CreateFail, err)
			return
		}
	}

	// 提交事务
	if err := session.Commit(); err != nil {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
