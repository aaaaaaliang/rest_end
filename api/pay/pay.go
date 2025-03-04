package pay

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

import (
	"context"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/go-pay/xlog"
)

// **支付宝支付接口**
func payHandler(c *gin.Context) {
	type Req struct {
		OrderCode string `json:"order_code" binding:"required" form:"order_code"`
	}

	var req Req
	if !utils.ValidationQuery(c, &req) {
		return
	}

	// 查询订单信息
	var order model.UserOrder
	has, err := config.DB.Where("code = ?", req.OrderCode).Get(&order)
	if err != nil || !has {
		response.Success(c, response.OrderNotFound, errors.New("订单不存在"))
		return
	}

	// 确保订单是未支付状态
	if order.Status != 5 {
		response.Success(c, response.OrderCanceled, errors.New("订单已支付或已取消"))
		return
	}

	// ✅ 初始化支付宝客户端
	client, err := alipay.NewClient(config.G.Pay.AppId, config.G.Pay.PrivateKey, config.G.Pay.IsProd)
	if err != nil {
		xlog.Error("初始化支付宝客户端失败:", err)
		response.Success(c, response.OrderPaymentFailed, errors.New("支付系统错误"))
		return
	}

	client.SetReturnUrl(config.G.Pay.SetReturnUrl)
	client.SetNotifyUrl(config.G.Pay.SetNotifyUrl)

	// ✅ 构造支付请求参数
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", order.Code).
		Set("total_amount", fmt.Sprintf("%.2f", order.TotalPrice)).
		Set("subject", "商品订单支付").
		Set("product_code", "FAST_INSTANT_TRADE_PAY")

	// ✅ 调用支付宝支付接口
	payURL, err := client.TradePagePay(context.Background(), bm)
	if err != nil {
		xlog.Error("TradePagePay 调用失败:", err)
		response.Success(c, response.OrderPaymentFailed, errors.New("支付请求失败"))
		return
	}

	// ✅ 返回支付链接
	response.SuccessWithData(c, response.SuccessCode, gin.H{"pay_url": payURL})
}
