package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
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

	// 设置连接读取超时
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

		// 如果是身份验证消息
		if msg.Type == "identify" {
			// 验证消息并设置角色
			if msg.Role == "user" {
				c.isSupport = false
			} else if msg.Role == "service" {
				c.isSupport = true

				c.hub.register <- c
			} else {
				errorMsg := []byte(`{"error": "无效的角色类型"}`)
				// 发给自己
				c.send <- errorMsg
				continue
			}
		} else if c.isSupport {
			// 如果是客服用户，直接处理消息
			msg.FromUser = c.userCode
			if msg.ToUser == "" {
				c.hub.broadcast <- msg
			} else {
				// 给目标用户发送消息
				if receiver, exists := c.hub.clients[msg.ToUser]; exists {
					receiver.send <- []byte(msg.Content)
				} else {
					log.Println("目标用户不存在，无法发送消息：", msg.ToUser)
				}
			}
		} else {
			// 普通用户消息处理
			if !c.hub.isSupportOnline() {
				errorMsg, _ := json.Marshal(map[string]string{
					"type":    "error",
					"content": "没有在线客服，无法发送消息",
				})
				c.send <- errorMsg
			} else {
				msg.FromUser = c.userCode
				// 普通用户发送消息给客服
				if msg.ToUser == "" {
					c.hub.broadcast <- msg
				} else {
					// 发送给所有在线客服
					for _, support := range c.hub.supportClients {
						msgBytes, _ := json.Marshal(msg)
						support.send <- msgBytes
					}
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
