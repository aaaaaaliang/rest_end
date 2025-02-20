package ws

import (
	"encoding/json"
	"log"
	"rest/model"
)

// Hub 维护 WebSocket 连接
type Hub struct {
	clients        map[string]*Client     // 连接的 WebSocket 客户端（user_code -> Client）
	register       chan *Client           // 连接注册通道
	unregister     chan *Client           // 断开连接通道
	broadcast      chan model.ChatMessage // 群聊消息
	privateMsg     chan model.ChatMessage // 私聊消息
	supportClients map[string]*Client     // 在线客服客户端
}

// NewHub 创建 WebSocket Hub
func NewHub() *Hub {
	h := &Hub{
		clients:        make(map[string]*Client),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan model.ChatMessage),
		privateMsg:     make(chan model.ChatMessage),
		supportClients: make(map[string]*Client), // 存储在线客服
	}

	go h.Run()

	return h
}
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// 注册用户到 Hub
			h.clients[client.userCode] = client
			if client.isSupport {
				// 如果是客服，添加到 supportClients 中
				h.supportClients[client.userCode] = client
				log.Println("客服用户在线:", client.userCode)
			}
			log.Println("当前在线用户数:", len(h.clients)) // 输出当前在线用户数

		case client := <-h.unregister:
			if _, ok := h.clients[client.userCode]; ok {
				// 删除连接的用户
				delete(h.clients, client.userCode)
				if client.isSupport {
					// 如果是客服，移除客服列表
					delete(h.supportClients, client.userCode)
					log.Println("客服用户下线:", client.userCode)
				}
				close(client.send)
			}

		case message := <-h.broadcast:
			log.Println("ws 1")
			// 处理广播消息
			for _, client := range h.clients {
				msgBytes, _ := json.Marshal(message)
				client.send <- msgBytes
			}
		case message := <-h.privateMsg:
			log.Println("ws 2")
			// 处理私聊消息
			if client, ok := h.clients[message.ToUser]; ok {
				msgBytes, _ := json.Marshal(message)
				client.send <- msgBytes
			}
		}
	}
}
