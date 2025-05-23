package dashboard

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"time"
)

func countTodayOrders(c *gin.Context) {
	// 获取今天开始/结束时间
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Unix()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local).Unix()

	// 获取过去 7 天的起始时间
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, time.Local).Unix()

	// 今天订单数量
	orderCount, err := config.DB.Table(model.UserOrder{}).
		Where("created >= ? AND created <= ?", startOfDay, endOfDay).
		Count()
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 今天用户数量
	userCount, err := config.DB.Table(model.Users{}).
		Where("created >= ? AND created <= ?", startOfDay, endOfDay).
		Count()
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 今天订单总金额
	totalAmount, err := config.DB.
		Where("created >= ? AND created <= ?", startOfDay, endOfDay).
		Sum(&model.UserOrder{}, "total_price")
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 热销商品统计：过去 7 天的 OrderDetail 聚合
	type ProductStat struct {
		ProductCode   string `json:"product_code"`
		ProductName   string `json:"product_name"`
		TotalQuantity int    `json:"quantity"`
	}

	var stats []ProductStat
	err = config.DB.Table("order_detail").
		Select("product_code, product_name, SUM(quantity) AS total_quantity").
		Where("created >= ?", startOfWeek).
		GroupBy("product_code, product_name").
		OrderBy("total_quantity DESC").
		Limit(10).
		Find(&stats)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 计算过去七天每天的销售金额
	dailySales := make(map[string]float64)
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -i)
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local).Unix()
		end := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 999999999, time.Local).Unix()

		dayAmount, err := config.DB.
			Where("created >= ? AND created <= ?", start, end).
			Sum(&model.UserOrder{}, "total_price")
		if err != nil {
			dayAmount = 0
		}
		dailySales[day.Format("2006-01-02")] = dayAmount
	}

	// 返回结构
	type Res struct {
		Orders      int64              `json:"orders"`
		Users       int64              `json:"users"`
		TotalAmount float64            `json:"total_amount"`
		Products    []ProductStat      `json:"products"`
		DailySales  map[string]float64 `json:"daily_sales"`
	}

	res := Res{
		Orders:      orderCount,
		Users:       userCount,
		TotalAmount: totalAmount,
		Products:    stats,
		DailySales:  dailySales,
	}

	response.SuccessWithData(c, response.SuccessCode, res)
}
