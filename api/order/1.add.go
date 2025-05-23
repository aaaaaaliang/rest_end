package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/state"
	"rest/utils"
	"time"
)

func addOrder(c *gin.Context) {
	type Req struct {
		TableNo    string              `json:"table_no" binding:"required"`
		Customer   string              `json:"customer"`
		TotalPrice float64             `json:"total_price" binding:"required"`
		Details    []model.OrderDetail `json:"details" binding:"required"`
		Remark     string              `json:"remark"`
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)
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
	log.Println("**************dsada", req.Details)

	// ✅ 1. 校验并锁定库存
	for _, detail := range req.Details {
		var product model.Products
		ok, err := session.Where("code = ?", detail.ProductCode).ForUpdate().Get(&product)
		if err != nil || !ok {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("商品不存在: %v", detail.ProductCode))
			return
		}
		if product.Count < int64(detail.Quantity) {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("商品 [%s] 库存不足", product.ProductsName))
			return
		}
	}

	// ✅ 2. 扣减库存（用安全 SQL 防止并发）
	for _, detail := range req.Details {
		sql := "UPDATE products SET count = count - ? WHERE code = ? AND count >= ?"
		res, err := session.Exec(sql, detail.Quantity, detail.ProductCode, detail.Quantity)
		if err != nil {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("扣减库存失败: %v", err))
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("库存不足，无法扣减: %s", detail.ProductCode))
			return
		}
	}

	// ✅ 3. 插入订单
	order := &model.UserOrder{
		UserCode:   userCode,
		UserName:   user.Username,
		TableNo:    req.TableNo,
		Customer:   req.Customer,
		TotalPrice: req.TotalPrice,
		Status:     int(state.OrderPendingPayment),
		Remark:     req.Remark,
		Version:    1,
	}
	order.Code = utils.GenerateOrderCode()
	order.Created = time.Now().Unix()

	if _, err := session.Insert(order); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// ✅ 4. 插入订单明细
	for _, detail := range req.Details {
		detailModel := model.OrderDetail{
			OrderCode:   order.Code,
			ProductCode: detail.ProductCode,
			ProductName: detail.ProductName,
			Quantity:    detail.Quantity,
			Price:       detail.Price,
			Picture:     detail.Picture,
			BasicModel:  model.BasicModel{Creator: userCode},
		}
		if _, err := session.Insert(&detailModel); err != nil {
			_ = session.Rollback()
			response.Success(c, response.ServerError, err)
			return
		}
	}

	// ✅ 5. 查询订单明细用于 ES
	var orderDetails []model.OrderDetail
	if err := session.Where("order_code = ?", order.Code).Find(&orderDetails); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// ✅ 6. 构造 ES 数据
	orderMap := map[string]interface{}{
		"code":         order.Code,
		"user_code":    order.UserCode,
		"user_name":    order.UserName,
		"table_no":     order.TableNo,
		"customer":     order.Customer,
		"total_price":  order.TotalPrice,
		"status":       order.Status,
		"remark":       order.Remark,
		"created":      order.Created,
		"order_detail": orderDetails,
	}

	log.Printf("🧩 写入ES的数据内容: %+v\n", orderMap)

	if err := saveOrderMapToES(orderMap); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, fmt.Errorf("存入ES失败: %v", err))
		return
	}

	// ✅ 7. 推送 MQ 消息
	message, err := json.Marshal(orderMap)
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

	// ✅ 8. 推送延时队列
	if err := publishDelayOrder(order); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	// ✅ 9. 提交事务
	if err := session.Commit(); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}

// 组装后的存 ES 方法
func saveOrderMapToES(order map[string]interface{}) error {
	esURL := fmt.Sprintf("%v/orders/_doc/%s", config.G.ES.Url, order["code"])
	jsonData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}
	req, _ := http.NewRequest("PUT", esURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("写入 Elasticsearch 失败: %v", err)
	}
	defer resp.Body.Close()
	log.Println("✅ 订单已写入 Elasticsearch:", order["code"], resp)
	return nil
}
