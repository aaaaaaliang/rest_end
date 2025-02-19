package order

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func addOrder(c *gin.Context) {
	type Req struct {
		TotalPrice float64             `json:"total_price" binding:"required"` // 订单总金额
		Details    []model.OrderDetail `json:"details" binding:"required"`     // 商品详情
		Remark     string              `json:"remark"`                         // 订单备注
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)

	// 开启事务
	session := config.DB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 构造订单模型
	order := &model.UserOrder{
		TotalPrice:  req.TotalPrice,
		Status:      1, // 订单状态 1已下单 2.制作中 3.已完成 4. 取消订单
		Remark:      req.Remark,
		OrderDetail: req.Details, // 直接使用切片
		UserCode:    userCode,
	}

	// 插入订单
	if affectRow, err := session.Insert(order); err != nil || affectRow != 1 {
		session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// 标记购物车中的相应商品为已下单
	for _, detail := range req.Details {
		// 假设 OrderDetail 中包含 ProductCode
		if detail.ProductCode == "" {
			session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("产品编号为空"))
			return
		}

		cart := model.UserCart{
			ProductCode: detail.ProductCode,
			UserCode:    userCode,
		}

		// 更新 IsOrdered 字段为 true
		affected, err := session.Where("user_code = ? AND product_code = ?", cart.UserCode, cart.ProductCode).
			Cols("is_ordered").
			Update(&model.UserCart{
				IsOrdered: true,
			})
		if err != nil {
			session.Rollback()
			response.Success(c, response.ServerError, err)
			return
		}
		if affected == 0 {
			session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("购物车项不存在或已下单"))
			return
		}
	}

	// 提交事务
	if err := session.Commit(); err != nil {
		session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
