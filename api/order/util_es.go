package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"rest/config"
	"rest/model"
)

// **存入 Elasticsearch**
func saveOrderToES(order *model.UserOrder) error {
	//"http://localhost:9200/orders/_doc/"
	url := fmt.Sprintf("%v/orders/_doc/", config.G.ES.Url)
	esURL := url + order.Code

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

	for _, detail := range order.OrderDetail {
		orderData["order_detail"] = append(orderData["order_detail"].([]map[string]interface{}), map[string]interface{}{
			"product_code": detail.ProductCode,
			"product_name": detail.ProductName,
			"quantity":     detail.Quantity,
			"price":        detail.Price,
			"picture":      detail.Picture,
		})
	}
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
