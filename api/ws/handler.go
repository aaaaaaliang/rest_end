package ws

//// **全局 WebSocket 升级器**
//var upgrader = websocket.Upgrader{
//	ReadBufferSize:  1024,
//	WriteBufferSize: 1024,
//	CheckOrigin: func(r *http.Request) bool {
//		return true // 允许所有来源连接
//	},
//}
//
//// **判断用户是否为客服**
//func isSupport(userCode string) bool {
//	// 查询用户的角色
//	exist, err := config.DB.Table(model.Users{}).Where("code = ? AND is_employee = 1", userCode).Exist()
//	if err != nil {
//		log.Println("isSupport 查询用户角色失败:", err)
//		return false
//	}
//	return exist
//}
//
//// ServeWs 处理 WebSocket 连接
//func ServeWs(hub *Hub, c *gin.Context) {
//	// 从 Cookie 获取 Token
//	token, err := c.Cookie("access_token")
//	if err != nil {
//		response.Success(c, response.Unauthorized, errors.New("未登录"))
//		return
//	}
//
//	// 解析 Token 获取 user_code
//	userCode, err := config.ParseJWT(token)
//	if err != nil {
//		response.Success(c, response.Unauthorized, errors.New("无效的 Token"))
//		return
//	}
//
//	// 判断用户是否为客服
//	isSupport := isSupport(userCode)
//
//	// 升级 HTTP 连接为 WebSocket
//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		log.Println("WebSocket 升级失败:", err)
//		return
//	}
//
//	// 创建 WebSocket 客户端
//	client := &Client{
//		hub:       hub,
//		conn:      conn,
//		send:      make(chan []byte, 256),
//		userCode:  userCode,
//		isSupport: isSupport, // 设置用户是否为客服
//	}
//
//	// 注册客户端到 WebSocket Hub
//	hub.register <- client
//
//	// 启动读/写协程
//	go client.writePump()
//	go client.readPump()
//}
//
//// 检查是否有在线客服
//func (h *Hub) isSupportOnline() bool {
//	// 返回支持的客服数量
//	return len(h.supportClients) > 0
//}
