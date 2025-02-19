package order

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

func updateOrder(c *gin.Context) {
	type Req struct {
		Code   string `json:"code" binding:"required"` // 订单编号
		Status int    `json:"status"`                  // 更新订单状态
		Remark string `json:"remark"`                  // 更新备注
		//TotalPrice float64             `json:"total_price"`             // 更新订单总金额
		//Details    []model.OrderDetail `json:"details"`                 // 更新订单商品详情
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	var o model.UserOrder
	if _, err := config.DB.Where("code = ?", req.Code).Get(&o); err != nil {
		response.Success(c, response.QueryFail, fmt.Errorf("updateOrder error: %v", err))
		return
	}
	if o.Status == 2 {
		response.Success(c, response.UpdateFail, errors.New("订单 已锁定"))
		return
	}

	// 构造更新模型
	order := &model.UserOrder{
		Status: req.Status,
		Remark: req.Remark,
		//TotalPrice:  req.TotalPrice,
		//OrderDetail: req.Details, // 直接使用切片
	}

	if affectRow, err := config.DB.Where("code = ?", req.Code).Update(order); err != nil || affectRow != 1 {
		response.Success(c, response.UpdateFail, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
