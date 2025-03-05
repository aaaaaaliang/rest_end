package ws

import (
	"encoding/json"
	"fmt"
	"rest/config"
	"rest/model"
	"strings"
	"time"
)

func fetchRecommendedDishes() (string, error) {
	var dishes []model.Products
	err := config.DB.Where("main = ?", 1).Limit(3).Find(&dishes)
	if err != nil {
		return "", err
	}

	// **如果没有推荐菜品**
	if len(dishes) == 0 {
		return "😢 当前暂无推荐菜品，请稍后再试~", nil
	}

	// **格式化推荐菜品**
	var dishList []string
	for _, dish := range dishes {
		dishInfo := fmt.Sprintf(
			"🍽 推荐菜品: %s\n💰 价格: ¥%.2f",
			dish.ProductsName,
			dish.Price,
		)
		dishList = append(dishList, dishInfo)
	}

	// **拼接推荐菜品**
	return strings.Join(dishList, "\n\n"), nil
}

// **获取用户订单**
func fetchOrders(userCode string) (string, error) {
	if userCode == "" {
		return "❌ 未找到用户信息，请先登录", nil
	}

	// **查询最近 5 笔订单**
	var orders []model.UserOrder
	err := config.DB.Where("user_code = ?", userCode).Limit(5).Desc("created").Find(&orders)
	if err != nil {
		return "", err
	}

	// **如果没有订单**
	if len(orders) == 0 {
		return "📌 您目前没有订单记录哦~", nil
	}

	// **格式化订单**
	var orderList []string

	for _, order := range orders {
		orderInfo := fmt.Sprintf(
			"📦 订单编号   : %s\n"+
				"🕒 下单时间   : %s\n"+
				"📜 状态       : %s\n"+
				"💰 总价       : %s\n"+
				"✏ 备注       : %s\n",
			order.Code,
			formatTimestamp(order.Created),
			getOrderStatus(order.Status),
			fmt.Sprintf("¥%.2f", order.TotalPrice),
			getRemark(order.Remark),
		)
		orderList = append(orderList, orderInfo)
	}

	// **提示用户查看更多**
	if len(orders) >= 5 {
		orderList = append(orderList, "📌 仅展示最近 5 笔订单，如需查看更多，请访问 '我的订单'")
	}

	// **拼接订单，确保订单之间有两个空行**
	return strings.Join(orderList, "\n\n"), nil
}

// **格式化时间**
func formatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "未知时间"
	}
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// **订单状态转换**
func getOrderStatus(status int) string {
	statusMap := map[int]string{
		1: "✅ 已下单",
		2: "⏳ 制作中",
		3: "🏁 已完成",
		4: "❌ 已取消",
		5: "💰 待支付",
	}
	return statusMap[status]
}

// **格式化备注**
func getRemark(remark string) string {
	if remark == "" {
		return "无备注"
	}
	return remark
}

// **检查字符串是否包含关键字**
func containsKeywords(text string, keywords []string) bool {
	text = strings.ToLower(text) // **转换为小写，提高匹配准确度**
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func sendAIResponse(c *Client, content string) {
	aiResponse, _ := json.Marshal(map[string]string{
		"type":      "chat",
		"content":   content,
		"from_user": "AI",
	})
	c.send <- aiResponse
}
