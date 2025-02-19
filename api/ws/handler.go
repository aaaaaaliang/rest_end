package ws

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"rest/config"
)

// **全局 WebSocket 升级器**
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源连接
	},
}

// ServeWs **处理 WebSocket 连接**
func ServeWs(hub *Hub, c *gin.Context) {
	// **从 Cookie 获取 Token**
	token, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// **解析 Token 获取 user_code**
	userCode, err := config.ParseJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token"})
		return
	}

	// **升级 HTTP 连接为 WebSocket**
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}

	// **创建 WebSocket 客户端**
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userCode: userCode,
	}

	// **注册客户端到 WebSocket Hub**
	hub.register <- client

	// **启动读/写协程**
	go client.writePump()
	go client.readPump()
}
