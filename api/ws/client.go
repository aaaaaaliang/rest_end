package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"rest/api/ai"
	"rest/model"
	"time"
)

// Client 代表 WebSocket 连接
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	userCode  string
	isSupport bool // 标记是否为客服
}

// 读取客户端消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("读取消息失败：", err)
			break
		}
		log.Println("收到消息:", string(message))

		var msg model.ChatMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("消息解析失败:", err)
			continue
		}

		// **身份验证**
		if msg.Type == "identify" {
			if msg.Role == "user" {
				c.isSupport = false
			} else if msg.Role == "service" {
				c.isSupport = true
				c.hub.register <- c
			} else {
				c.send <- []byte(`{"error": "无效的角色类型"}`)
				continue
			}
		} else if c.isSupport {
			// **人工客服**
			msg.FromUser = c.userCode
			if msg.ToUser == "" {
				c.hub.broadcast <- msg
			} else {
				if receiver, exists := c.hub.clients[msg.ToUser]; exists {
					receiver.send <- []byte(msg.Content)
				} else {
					log.Println("目标用户不存在，无法发送消息：", msg.ToUser)
				}
			}
		} else {
			// **普通用户发送消息**
			if !c.hub.isSupportOnline() {
				// **调用 AI 客服**
				aiReply, err := ai.SendToAI("deepseek-r1:1.5b", msg.Content) // 使用本地 AI 方法
				if err != nil {
					errorMsg, _ := json.Marshal(map[string]string{
						"type":    "error",
						"content": "AI 客服暂时不可用",
					})
					c.send <- errorMsg
				} else {
					aiResponse, _ := json.Marshal(map[string]string{
						"type":      "chat",
						"content":   aiReply,
						"from_user": "AI",
					})
					c.send <- aiResponse
				}
			} else {
				// **人工客服在线，消息转发给客服**
				msg.FromUser = c.userCode
				for _, support := range c.hub.supportClients {
					msgBytes, _ := json.Marshal(msg)
					support.send <- msgBytes
				}
			}
		}
	}
}

// writePump 监听 Hub 广播并发送给客户端
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
