package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
	"time"
)

func updateOrder(c *gin.Context) {
	type Req struct {
		Code    string `json:"code" binding:"required"` // 订单编号
		Status  int    `json:"status"`                  // 更新订单状态
		Remark  string `json:"remark"`                  // 更新备注
		Version int    `json:"version"`                 // 乐观锁版本号
	}

	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}

	var o model.UserOrder
	// 查找订单
	if _, err := config.DB.Where("code = ?", req.Code).Get(&o); err != nil {
		response.Success(c, response.QueryFail, fmt.Errorf("updateOrder error: %v", err))
		return
	}

	// 允许的状态变更映射
	validTransitions := map[int][]int{
		1: {1, 2}, // 1 可以变成 1 或 2
		2: {2, 3}, // 2 可以变成 2 或 3
		3: {3},
		4: {4},
		5: {5, 4}, // 5 可以变成 5 或 4
	}

	// 检查是否允许该状态变更
	allowed, exists := validTransitions[o.Status]
	if !exists || !contains(allowed, req.Status) {
		response.Success(c, response.UpdateFail, errors.New("订单状态不可更改"))
		return
	}

	// 检查版本号是否一致
	if o.Version != req.Version {
		response.Success(c, response.UpdateFail, errors.New("数据已被其他操作修改"))
		return
	}

	// 开启事务
	session := config.DB.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 构造更新模型
	o.Status = req.Status
	o.Remark = req.Remark
	o.Version += 1
	o.Updated = time.Now().Unix()

	// 执行更新操作
	if affectRow, err := session.Where("code = ?", req.Code).Update(&o); err != nil || affectRow != 1 {
		_ = session.Rollback()
		response.Success(c, response.UpdateFail, err)
		return
	}

	// **同步更新到 Elasticsearch**
	if err := updateOrderInES(&o); err != nil {
		_ = session.Rollback()
		response.Success(c, response.UpdateFail, fmt.Errorf("更新 Elasticsearch 失败: %v", err))
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

// contains 判断切片中是否包含某个值
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// **更新 Elasticsearch 中的订单**
func updateOrderInES(order *model.UserOrder) error {
	esURL := "http://localhost:9200/orders/_update/" + order.Code + "?refresh=wait_for"
	//esURL := "http://localhost:9200/orders/_update/" + order.Code

	// 构造更新数据
	updateData := map[string]interface{}{
		"doc": map[string]interface{}{
			"status":  order.Status,
			"remark":  order.Remark,
			"updated": order.Updated,
			"version": order.Version,
		},
	}

	// 转换为 JSON
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}

	// 发送请求
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
