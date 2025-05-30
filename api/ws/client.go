package ws

// Client 代表 WebSocket 连接
//type Client struct {
//	hub       *Hub
//	conn      *websocket.Conn
//	send      chan []byte
//	userCode  string
//	isSupport bool // 标记是否为客服
//}

//func (c *Client) readPump() {
//	defer func() {
//		c.hub.unregister <- c
//		_ = c.conn.Close()
//	}()
//
//	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
//	c.conn.SetPongHandler(func(string) error {
//		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
//		return nil
//	})
//
//	for {
//		_, message, err := c.conn.ReadMessage()
//		if err != nil {
//			logger.Println("读取消息失败：", err)
//			break
//		}
//		logger.Println("收到消息:", string(message))
//
//		var msg model.ChatMessage
//		if err := json.Unmarshal(message, &msg); err != nil {
//			logger.Println("消息解析失败:", err)
//			continue
//		}
//
//		// **身份验证**
//		if msg.Type == "identify" {
//			if msg.Role == "user" {
//				c.isSupport = false
//			} else if msg.Role == "service" {
//				c.isSupport = true
//				c.hub.register <- c
//			} else {
//				c.send <- []byte(`{"error": "无效的角色类型"}`)
//				continue
//			}
//		} else if c.isSupport {
//			// **人工客服**
//			msg.FromUser = c.userCode
//			if msg.ToUser == "" {
//				logger.Println("客服在线，广播消息给所有人")
//				c.hub.broadcast <- msg
//			} else {
//				if receiver, exists := c.hub.clients[msg.ToUser]; exists {
//					logger.Println("客服发送消息给目标用户", msg.ToUser)
//					receiver.send <- []byte(msg.Content)
//				} else {
//					logger.Println("目标用户不存在，无法发送消息：", msg.ToUser)
//				}
//			}
//		} else {
//			// **普通用户发送消息**
//			go c.handleUserRequest(msg)
//		}
//	}
//}
//
//func (c *Client) handleUserRequest(msg model.ChatMessage) {
//	var responseMsg string
//
//	// **订单查询**
//	if containsKeywords(msg.Content, []string{"查看订单", "我的订单", "订单"}) {
//		orders, err := fetchOrders(c.userCode)
//		if err != nil {
//			responseMsg = "❌ 获取订单失败，请稍后再试"
//		} else {
//			responseMsg = orders
//		}
//		sendAIResponse(c, responseMsg)
//		return
//	}
//
//	// **推荐菜品**
//	if containsKeywords(msg.Content, []string{"推荐菜", "特色菜", "推荐几道菜"}) {
//		dishes, err := fetchRecommendedDishes()
//		if err != nil {
//			responseMsg = "❌ 获取推荐菜品失败，请稍后再试"
//		} else {
//			responseMsg = dishes
//		}
//		sendAIResponse(c, responseMsg)
//		return
//	}
//
//	// **调用 AI**
//	aiReply, err := ai.SendToAI("deepseek-r1:1.5b", msg.Content)
//	if err != nil {
//		responseMsg = "❌ AI 客服暂时不可用"
//	} else {
//		responseMsg = aiReply
//	}
//	sendAIResponse(c, responseMsg)
//}
//
//// writePump 监听 Hub 广播并发送给客户端
//func (c *Client) writePump() {
//	ticker := time.NewTicker(54 * time.Second)
//	defer func() {
//		ticker.Stop()
//		_ = c.conn.Close()
//	}()
//	for {
//		select {
//		case message, ok := <-c.send:
//			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
//			if !ok {
//				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
//				return
//			}
//			err := c.conn.WriteMessage(websocket.TextMessage, message)
//			if err != nil {
//				return
//			}
//		case <-ticker.C:
//			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
//			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
//				return
//			}
//		}
//	}
//}
