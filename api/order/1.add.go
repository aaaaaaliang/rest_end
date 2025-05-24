package order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/state"
	"rest/utils"
)

func addOrder(c *gin.Context) {
	type Req struct {
		TableNo             string              `json:"table_no" binding:"required"`
		Customer            string              `json:"customer"`
		TotalPrice          float64             `json:"total_price" binding:"required"`
		Details             []model.OrderDetail `json:"details" binding:"required"`
		Remark              string              `json:"remark"`
		CouponCode          string              `json:"coupon_code"`           // 新增：优惠券 code（user_coupon 表）
		ClientPayableAmount float64             `json:"client_payable_amount"` // 🆕 前端传入的最终支付金额
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

	for _, detail := range req.Details {
		res, err := session.Exec("UPDATE products SET count = count - ? WHERE code = ? AND count >= ?", detail.Quantity, detail.ProductCode, detail.Quantity)
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

	// ➤ 处理优惠券
	var couponAmount float64 = 0
	if req.CouponCode != "" {
		var userCoupon model.UserCoupon
		has, err := session.Where("code = ? AND user_code = ? AND status = 0", req.CouponCode, userCode).Get(&userCoupon)
		if err != nil || !has {
			_ = session.Rollback()
			response.Success(c, response.BadRequest, fmt.Errorf("优惠券无效或已使用 %v", err))
			return
		}
		if userCoupon.ExpireTime < time.Now().Unix() {
			_ = session.Rollback()
			response.Success(c, response.BadRequest, fmt.Errorf("优惠券已过期"))
			return
		}

		var tpl model.CouponTemplate
		hasTpl, err := session.Where("code = ?", userCoupon.TemplateCode).Get(&tpl)
		if err != nil || !hasTpl {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("优惠券模板不存在"))
			return
		}

		if tpl.Type == "full" && req.TotalPrice >= tpl.MinAmount {
			couponAmount = tpl.Quota
		} else if tpl.Type == "discount" && req.TotalPrice >= tpl.MinAmount {
			couponAmount = req.TotalPrice * (1 - tpl.Quota)
		} else if tpl.Type == "cash" {
			couponAmount = tpl.Quota
		}
		if couponAmount > req.TotalPrice {
			couponAmount = req.TotalPrice
		}

		userCoupon.Status = 1
		userCoupon.UseTime = time.Now().Unix()
		if _, err := session.ID(userCoupon.Id).Update(&userCoupon); err != nil {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("更新优惠券状态失败"))
			return
		}
	}

	// 校验客户端传入的应付金额与后端计算是否一致（防止作弊）
	realPayable := req.TotalPrice - couponAmount
	realPayable = math.Round(realPayable*100) / 100 // 四舍五入保留两位小数
	if req.ClientPayableAmount != 0 && math.Abs(realPayable-req.ClientPayableAmount) > 0.01 {
		_ = session.Rollback()
		response.Success(c, response.BadRequest, fmt.Errorf("订单金额不一致，客户端: %.2f，服务端: %.2f", req.ClientPayableAmount, realPayable))
		return
	}

	order := &model.UserOrder{
		UserCode:     userCode,
		UserName:     user.Username,
		TableNo:      req.TableNo,
		Customer:     req.Customer,
		TotalPrice:   req.TotalPrice,
		Status:       int(state.OrderPendingPayment),
		Remark:       req.Remark,
		Version:      1,
		CouponAmount: req.TotalPrice - couponAmount,
		CouponCode:   req.CouponCode,
	}
	order.Code = utils.GenerateOrderCode()
	if _, err := session.Insert(order); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

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

	var orderDetails []model.OrderDetail
	if err := session.Where("order_code = ?", order.Code).Find(&orderDetails); err != nil {
		_ = session.Rollback()
		response.Success(c, response.ServerError, err)
		return
	}

	var orderDetailsForES []map[string]interface{}
	for _, d := range orderDetails {
		orderDetailsForES = append(orderDetailsForES, map[string]interface{}{
			"product_code": d.ProductCode,
			"product_name": d.ProductName,
			"quantity":     d.Quantity,
			"price":        d.Price,
			"picture":      d.Picture,
		})
	}

	orderMap := map[string]interface{}{
		"code":          order.Code,
		"user_code":     order.UserCode,
		"user_name":     order.UserName,
		"table_no":      order.TableNo,
		"customer":      order.Customer,
		"total_price":   order.TotalPrice,
		"coupon_amount": order.CouponAmount,
		"status":        order.Status,
		"remark":        order.Remark,
		"created":       time.Now().Unix(),
		"order_detail":  orderDetailsForES,
	}

	err := saveOrderMapToES(order.Code, orderMap)
	if err != nil {
		log.Printf("❌ saveOrderMapToES 返回错误: %v\n", err)
		_ = session.Rollback()
		response.Success(c, response.ServerError, fmt.Errorf("存入ES失败: %v", err))
		return
	}

	productCodes := make([]interface{}, 0)
	for _, d := range req.Details {
		productCodes = append(productCodes, d.ProductCode)
	}

	if len(productCodes) > 0 {
		if _, err := session.In("product_code", productCodes...).
			And("user_code = ?", userCode).
			Cols("is_ordered").
			Update(&model.UserCart{IsOrdered: true}); err != nil {
			_ = session.Rollback()
			response.Success(c, response.ServerError, fmt.Errorf("标记购物车为已下单失败: %v", err))
			return
		}
	}

	if err := session.Commit(); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	response.Success(c, response.SuccessCode)
}

func saveOrderMapToES(docID string, order map[string]interface{}) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %v", err)
	}

	res, err := config.ESClient.Index(
		"orders",
		bytes.NewReader(jsonData),
		config.ESClient.Index.WithDocumentID(docID),
		config.ESClient.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("写入 Elasticsearch 失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		log.Printf("❌ Elasticsearch 错误响应，状态码: %s，响应体: %s\n", res.Status(), string(body))
		return fmt.Errorf("写入 Elasticsearch 错误响应: %s", res.Status())
	}

	log.Println("✅ 订单已写入 Elasticsearch:", docID)
	return nil
}
