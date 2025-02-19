package dashboard

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"sort"
	"time"
)

func countTodayOrders(c *gin.Context) {
	// 获取今天的开始时间和结束时间
	startOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Unix()
	endOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 999999999, time.Local).Unix()

	// 获取过去七天的开始时间
	startOfWeek := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-6, 0, 0, 0, 0, time.Local).Unix()

	// 获取今天的订单数量
	orderCount, err := config.DB.Where("created >= ? AND created <= ?", startOfDay, endOfDay).Table(model.UserOrder{}).Count()
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 获取今天的用户数量
	userCount, err := config.DB.Where("created >= ? AND created <= ?", startOfDay, endOfDay).Table(model.Users{}).Count()
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 计算今天的所有订单的总金额
	totalAmount, err := config.DB.Where("created >= ? AND created <= ?", startOfDay, endOfDay).Sum(&model.UserOrder{}, "total_price")
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 查询过去七天的所有订单
	var orders []model.UserOrder
	err = config.DB.Where("created >= ? AND created <= ?", startOfWeek, endOfDay).Find(&orders)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 用来统计产品销量（过去七天）
	productSales := make(map[string]struct {
		ProductName string `json:"product_name"`
		Quantity    int    `json:"quantity"`
	})

	// 遍历订单，解析 OrderDetail 统计菜品销售情况
	for _, order := range orders {
		for _, detail := range order.OrderDetail {
			if product, exists := productSales[detail.ProductCode]; exists {
				productSales[detail.ProductCode] = struct {
					ProductName string `json:"product_name"`
					Quantity    int    `json:"quantity"`
				}{
					ProductName: detail.ProductName,
					Quantity:    product.Quantity + detail.Quantity,
				}
			} else {
				productSales[detail.ProductCode] = struct {
					ProductName string `json:"product_name"`
					Quantity    int    `json:"quantity"`
				}{
					ProductName: detail.ProductName,
					Quantity:    detail.Quantity,
				}
			}
		}
	}

	// 将统计结果转换为切片
	type Result struct {
		ProductCode string `json:"product_code"`
		ProductName string `json:"product_name"`
		Quantity    int    `json:"quantity"`
	}

	var productRanking []Result
	for code, sales := range productSales {
		productRanking = append(productRanking, Result{
			ProductCode: code,
			ProductName: sales.ProductName,
			Quantity:    sales.Quantity,
		})
	}

	// 按销量降序排序
	sort.Slice(productRanking, func(i, j int) bool {
		return productRanking[i].Quantity > productRanking[j].Quantity
	})

	// 计算过去七天每天的销售金额
	dailySales := make(map[string]float64)

	// 计算每天的销售金额（过去六天+今天）
	for i := 0; i <= 6; i++ {
		dayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-i, 0, 0, 0, 0, time.Local).Unix()
		dayEnd := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-i, 23, 59, 59, 999999999, time.Local).Unix()

		// 计算每天的总金额
		dailyAmount, err := config.DB.Where("created >= ? AND created <= ?", dayStart, dayEnd).Sum(&model.UserOrder{}, "total_price")
		if err != nil {
			dailyAmount = 0
		}

		// 设置日期作为 key
		dailySales[time.Unix(dayStart, 0).Format("2006-01-02")] = dailyAmount
	}

	// 构造返回结果
	type Res struct {
		Orders      int64              `json:"orders"`
		Users       int64              `json:"users"`
		TotalAmount float64            `json:"total_amount"`
		Products    []Result           `json:"products"`
		DailySales  map[string]float64 `json:"daily_sales"`
	}

	// **只改 Products 赋值，其他不变**
	res := Res{
		Orders:      orderCount,
		Users:       userCount,
		TotalAmount: totalAmount,
		Products:    productRanking, // **这里返回过去七天的热销菜品**
		DailySales:  dailySales,
	}

	// 返回统计数据
	response.SuccessWithData(c, response.SuccessCode, res)
}
