package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func updateOrder(c *gin.Context) {
	type Req struct {
		Code   string `json:"code" binding:"required"` // 订单编号
		Status int    `json:"status"`                  // 更新订单状态
		Remark string `json:"remark"`                  // 更新备注
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	session := config.DB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	var o model.UserOrder
	if _, err := session.Where("code = ?", req.Code).ForUpdate().Get(&o); err != nil {
		_ = session.Rollback()
		response.Success(c, response.QueryFail, fmt.Errorf("查询订单失败: %v", err))
		return
	}

	if o.Code == "" {
		_ = session.Rollback()
		response.Success(c, response.UpdateFail, errors.New("订单不存在"))
		return
	}

	// 状态变更规则
	validTransitions := map[int][]int{
		1: {1, 2, 4}, // 允许从未支付变为制作中、已取消
		2: {2, 3},    // 制作中只能变完成
		3: {3},       // 完成后不可改
		4: {4},       // 已取消不可改
		5: {5, 4, 1}, // 备用状态逻辑
	}

	allowed, exists := validTransitions[o.Status]
	if !exists || !contains(allowed, req.Status) {
		_ = session.Rollback()
		response.Success(c, response.UpdateFail, errors.New("订单状态不可更改"))
		return
	}

	// 如果是从 待支付（1） → 已取消（4） 则执行库存回补
	if o.Status == 1 && req.Status == 4 {
		var details []model.OrderDetail
		if err := session.Where("order_code = ?", o.Code).Find(&details); err != nil {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("查询订单明细失败: %v", err))
			return
		}

		for _, d := range details {
			_, err := session.Exec("UPDATE products SET count = count + ? WHERE code = ?", d.Quantity, d.ProductCode)
			if err != nil {
				_ = session.Rollback()
				response.Success(c, response.UpdateFail, fmt.Errorf("库存回补失败: %v", err))
				return
			}
		}
	}

	// 更新订单状态和备注
	o.Status = req.Status
	o.Remark = req.Remark
	o.Updated = time.Now().Unix()

	if affectRow, err := session.Where("code = ?", req.Code).Update(&o); err != nil || affectRow != 1 {
		_ = session.Rollback()
		response.Success(c, response.UpdateFail, err)
		return
	}

	if err := updateOrderInES(&o); err != nil {
		_ = session.Rollback()
		response.Success(c, response.UpdateFail, fmt.Errorf("更新 Elasticsearch 失败: %v", err))
		return
	}

	if err := session.Commit(); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}

// 判断切片是否包含元素
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// 更新订单状态至 Elasticsearch
func updateOrderInES(order *model.UserOrder) error {
	esURL := fmt.Sprintf("%s/orders/_update/%s?refresh=wait_for", config.G.ES.Url, order.Code)
	updateData := map[string]interface{}{
		"doc": map[string]interface{}{
			"status":  order.Status,
			"remark":  order.Remark,
			"updated": order.Updated,
			"version": order.Version,
		},
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}

	req, _ := http.NewRequest("POST", esURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("更新 Elasticsearch 失败: %v", err)
	}
	defer resp.Body.Close()

	log.Println("✅ 订单已更新到 Elasticsearch:", order.Code)
	return nil
}
