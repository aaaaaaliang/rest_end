package order

import (
	"encoding/json"
	"log"
	"rest/state"
	"time"

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

	var user model.Users
	if _, err := config.DB.Where("code = ?", userCode).Get(&user); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}
	log.Println("用户信息:", user.Code, user.Username)

	// 构造订单模型
	order := &model.UserOrder{
		TotalPrice:  req.TotalPrice,
		Status:      int(state.OrderPendingPayment), // 订单状态：待支付   // 订单状态 1.已下单 2.制作中 3.已完成 4. 取消订单 5.待支付
		Remark:      req.Remark,
		OrderDetail: req.Details,
		UserCode:    userCode,
		UserName:    user.Username,
	}
	order.Code = utils.GenerateOrderCode()
	order.Created = time.Now().Unix()
	if affectRow, err := session.Insert(order); err != nil || affectRow != 1 {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	message, err := json.Marshal(order)
	if err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	if err := publishMessage("order_queue", message); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// 2. 发布延时消息到延时队列，用于30分钟后订单超时检测
	if err := publishDelayOrder(order); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	////存入 Elasticsearch
	//if err := saveOrderToES(order); err != nil {
	//	_ = session.Rollback()
	//	response.Success(c, response.ServerError, err)
	//	return
	//}

	// 提交事务
	if err := session.Commit(); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}
