package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		Status:      5, // 订单状态：待支付   // 订单状态  待支付 1已下单 2.制作中 3.已完成 4. 取消订单 5.待支付
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

	//存入 Elasticsearch
	if err := saveOrderToES(order); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// 提交事务
	if err := session.Commit(); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}

// **存入 Elasticsearch**
func saveOrderToES(order *model.UserOrder) error {
	esURL := "http://localhost:9200/orders/_doc/" + order.Code

	orderData := map[string]interface{}{
		"code":         order.Code,
		"user_code":    order.UserCode,
		"user_name":    order.UserName,
		"total_price":  order.TotalPrice,
		"status":       order.Status,
		"remark":       order.Remark,
		"created":      order.Created,
		"order_detail": []map[string]interface{}{},
	}

	// 转成nested
	for _, detail := range order.OrderDetail {
		orderData["order_detail"] = append(orderData["order_detail"].([]map[string]interface{}), map[string]interface{}{
			"product_code": detail.ProductCode,
			"product_name": detail.ProductName,
			"quantity":     detail.Quantity,
			"price":        detail.Price,
			"picture":      detail.Picture,
		})
	}

	// 转换为 JSON
	jsonData, err := json.Marshal(orderData)
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}

	// 发送请求
	req, _ := http.NewRequest("POST", esURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("写入 Elasticsearch 失败: %v", err)
	}
	defer resp.Body.Close()

	log.Println("✅ 订单已存入 Elasticsearch:", order.Code)
	return nil
}
