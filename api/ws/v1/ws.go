package handler

//
//type WsMessage struct {
//	SessionCode string `json:"session_code"` // 会话唯一标识
//	SenderCode  string `json:"sender_code"`  // 用户code（后端注入，不需要前端传）
//	SenderType  string `json:"sender_type"`  // "customer" or "agent"
//	Content     string `json:"content"`      // 聊天内容
//}
//
//type ChatSessionConn struct {
//	CustomerConn *websocket.Conn
//	AgentConn    *websocket.Conn
//}
//
//var (
//	WSConnPool = make(map[string]*ChatSessionConn)
//	connLock   sync.RWMutex
//	logger     = logger.New("chat")
//)
//
//// todo
//var upgrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}
//
//func chatWebSocket(c *gin.Context) {
//	userCode := utils.GetUser(c)
//	sessionCode := c.Query("session_code")
//	userType := c.Query("user_type") // "customer" or "agent"：可后期从角色系统判断
//
//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		logger.WithFields(map[string]interface{}{
//			"session_code": sessionCode,
//			"user_type":    userType,
//			"user_code":    userCode,
//		}).Errorf("WebSocket 升级失败: %v", err)
//		return
//	}
//	defer func(conn *websocket.Conn) {
//		err := conn.Close()
//		if err != nil {
//			config.Log.Errorf("WebSocket conn %v", err)
//		}
//		tryCloseSession(sessionCode) // ✅ 增加这一行
//
//	}(conn)
//
//	// 加入连接池
//	connLock.Lock()
//	if WSConnPool[sessionCode] == nil {
//		WSConnPool[sessionCode] = &ChatSessionConn{}
//	}
//	if userType == "customer" {
//		WSConnPool[sessionCode].CustomerConn = conn
//	} else {
//		WSConnPool[sessionCode].AgentConn = conn
//	}
//	connLock.Unlock()
//
//	logger.WithFields(map[string]interface{}{
//		"session_code": sessionCode,
//		"user_type":    userType,
//		"user_code":    userCode,
//	}).Info("用户已加入会话")
//
//	for {
//		var msg WsMessage
//		err := conn.ReadJSON(&msg)
//		if err != nil {
//			logger.WithField("session_code", sessionCode).
//				Errorf("读取消息失败: %v", err)
//			break
//		}
//
//		// 注入身份信息，避免客户端伪造 sender_code
//		msg.SenderCode = userCode
//		msg.SenderType = userType
//
//		// 写入数据库
//		chat := model.ChatMessage{
//			SessionCode: msg.SessionCode,
//			SenderCode:  msg.SenderCode,
//			SenderType:  msg.SenderType,
//			Content:     msg.Content,
//		}
//		if _, err := config.DB.Insert(&chat); err != nil {
//			logger.WithField("session_code", msg.SessionCode).
//				Errorf("消息入库失败: %v", err)
//		}
//
//		// 转发给对方
//		connLock.RLock()
//		sessionConn := WSConnPool[sessionCode]
//		var receiver *websocket.Conn
//		if msg.SenderType == "customer" {
//			receiver = sessionConn.AgentConn
//		} else {
//			receiver = sessionConn.CustomerConn
//		}
//		connLock.RUnlock()
//
//		if receiver != nil {
//			if err := receiver.WriteJSON(msg); err != nil {
//				logger.WithFields(map[string]interface{}{
//					"session_code": sessionCode,
//					"to":           msg.SenderType,
//				}).Errorf("发送消息给对方失败: %v", err)
//			}
//		}
//	}
//}
//
//// 自动判断是否关闭会话
//func tryCloseSession(sessionCode string) {
//	connLock.Lock()
//	defer connLock.Unlock()
//
//	sessionConn := WSConnPool[sessionCode]
//	if sessionConn == nil {
//		return
//	}
//
//	bothClosed := true
//	if sessionConn.CustomerConn != nil {
//		if err := sessionConn.CustomerConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second)); err == nil {
//			bothClosed = false
//		}
//	}
//	if sessionConn.AgentConn != nil {
//		if err := sessionConn.AgentConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second)); err == nil {
//			bothClosed = false
//		}
//	}
//
//	if bothClosed {
//		delete(WSConnPool, sessionCode)
//		// 更新数据库为 ended
//		_, err := config.DB.Table(model.ChatSession{}).
//			Where("code = ? AND status = 'active'", sessionCode).
//			Update(map[string]interface{}{"status": "ended"})
//		if err != nil {
//			logger.WithField("session_code", sessionCode).
//				Errorf("会话结束失败: %v", err)
//		} else {
//			logger.WithField("session_code", sessionCode).
//				Info("会话已结束")
//		}
//	}
//}
