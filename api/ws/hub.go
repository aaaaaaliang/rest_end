package ws

import (
	"context"
	"encoding/json"
	"log"
	"rest/config"
	"rest/model"
	"time"
)

// Hub 维护 WebSocket 连接
type Hub struct {
	clients      map[string]*Client // 连接的 WebSocket 客户端（user_code -> Client）
	register     chan *Client       // 连接注册通道
	unregister   chan *Client       // 断开连接通道
	broadcast    chan ChatMessage   // 群聊消息
	privateMsg   chan ChatMessage   // 私聊消息
	redisChannel string             // Redis 频道
}

type ChatMessage struct {
	FromUser  string `json:"from_user"`
	ToUser    string `json:"to_user,omitempty"` // **空值省略**
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // text / image / file
}

// NewHub 创建 WebSocket Hub
func NewHub(ctx context.Context, redisChannel string) *Hub {
	h := &Hub{
		clients:      make(map[string]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan ChatMessage),
		privateMsg:   make(chan ChatMessage),
		redisChannel: redisChannel,
	}

	go h.Run()
	go h.subscribeRedis(ctx, redisChannel)

	return h
}

// Run 监听 WebSocket 事件
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userCode] = client
			log.Println("新客户端注册：", client.userCode)
			h.sendOfflineMessages(client.userCode)

		case client := <-h.unregister:
			if _, ok := h.clients[client.userCode]; ok {
				delete(h.clients, client.userCode)
				close(client.send)
				log.Println("客户端注销：", client.userCode)
			}

		case msg := <-h.broadcast:
			h.storeMessage(msg)
			msgJSON, _ := json.Marshal(msg) // ✅ 发送完整 JSON
			for _, client := range h.clients {
				client.send <- msgJSON
			}

		case msg := <-h.privateMsg:
			h.storeMessage(msg)
			if receiver, exists := h.clients[msg.ToUser]; exists {
				msgJSON, _ := json.Marshal(msg)
				receiver.send <- msgJSON
			} else {
				cacheOfflineMessage(msg.ToUser, msg) // ✅ 传递整个 ChatMessage 结构体
			}

		}
	}
}

// 存储聊天记录
func (h *Hub) storeMessage(msg ChatMessage) {
	chat := model.ChatMessage{
		FromUser:  msg.FromUser,
		ToUser:    msg.ToUser,
		Content:   msg.Content,
		Timestamp: time.Unix(msg.Timestamp, 0),
		Type:      msg.Type,
	}
	_, err := config.DB.Insert(&chat)
	if err != nil {
		log.Println("存储聊天记录失败:", err)
	}
}

func (h *Hub) sendOfflineMessages(userCode string) {
	messages := getOfflineMessages(userCode)
	if messages == nil {
		return
	}
	for _, msgStr := range messages {
		var msg ChatMessage
		if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
			log.Println("解析离线消息失败:", err)
			continue
		}
		if client, exists := h.clients[userCode]; exists {
			msgJSON, _ := json.Marshal(msg)
			client.send <- msgJSON
		}
	}
	deleteOfflineMessages(userCode)
}
