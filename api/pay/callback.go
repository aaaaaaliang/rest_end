package pay

import (
	"github.com/gin-gonic/gin"
	"github.com/go-pay/xlog"
	"net/http"
	"rest/config"
	"rest/model"
	"time"
)

// **支付宝支付回调**
func payNotifyHandler(c *gin.Context) {
	// 获取支付宝回调的参数
	orderCode := c.PostForm("out_trade_no")   // 订单号
	tradeStatus := c.PostForm("trade_status") // 交易状态

	// 日志记录
	xlog.Info("收到支付宝回调: 订单号:", orderCode, "交易状态:", tradeStatus)

	// **订单号不能为空**
	if orderCode == "" {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// **查询订单**
	var order model.UserOrder
	has, err := config.DB.Where("code = ?", orderCode).Get(&order)
	if err != nil || !has {
		xlog.Error("订单不存在:", orderCode)
		c.String(http.StatusOK, "success") // 避免支付宝继续重试
		return
	}

	// **幂等处理：如果订单已支付，直接返回 success**
	if order.Status == 1 {
		xlog.Info("订单已支付，跳过:", orderCode)
		c.String(http.StatusOK, "success")
		return
	}

	// **判断支付成功**
	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		// **更新订单状态为已支付**
		order.Status = 1
		order.Updated = time.Now().Unix()
		_, err = config.DB.ID(order.Id).Cols("status", "updated").Update(&order)
		if err != nil {
			xlog.Error("更新订单状态失败:", err)
			c.String(http.StatusInternalServerError, "fail")
			return
		}
		xlog.Info("订单支付成功，状态已更新:", orderCode)
	}

	// **支付宝回调必须返回 success，否则会继续重试**
	c.String(http.StatusOK, "success")
}
