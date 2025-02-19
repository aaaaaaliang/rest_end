package ws

import (
	"context"
	"encoding/json"
	"log"
	"rest/config"
	"time"
)

func publishMessage(channel string, message ChatMessage) {
	// 确保消息有时间戳
	if message.Timestamp == 0 {
		message.Timestamp = time.Now().Unix()
	}

	msgJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("消息序列化失败:", err)
		return
	}
	err = config.R.Publish(context.Background(), channel, string(msgJSON)).Err()
	if err != nil {
		log.Println("Redis 发布消息失败:", err)
	}
}

func (h *Hub) subscribeRedis(ctx context.Context, channel string) {
	pubsub := config.R.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Println("Redis 订阅取消，退出监听")
			return
		case msg := <-ch:
			var chatMsg ChatMessage
			if err := json.Unmarshal([]byte(msg.Payload), &chatMsg); err != nil {
				log.Println("解析 Redis 消息失败:", err)
				continue
			}

			// **区分群聊和私聊**
			if chatMsg.ToUser == "" {
				h.broadcast <- chatMsg
			} else {
				h.privateMsg <- chatMsg
			}
		}
	}
}

func cacheOfflineMessage(userCode string, message ChatMessage) {
	// 确保消息有时间戳
	if message.Timestamp == 0 {
		message.Timestamp = time.Now().Unix()
	}

	key := "offline_messages:" + userCode
	msgJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("序列化离线消息失败:", err)
		return
	}

	err = config.R.RPush(context.Background(), key, string(msgJSON)).Err()
	if err != nil {
		log.Println("存储离线消息失败:", err)
	}
	config.R.Expire(context.Background(), key, 7*24*time.Hour)
}

// **获取离线消息**
func getOfflineMessages(userCode string) []string {
	key := "offline_messages:" + userCode
	messages, err := config.R.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		log.Println("获取离线消息失败:", err)
		return nil
	}
	return messages
}

// **删除离线消息**
func deleteOfflineMessages(userCode string) {
	key := "offline_messages:" + userCode
	messages := getOfflineMessages(userCode)
	if len(messages) == 0 {
		return
	}

	err := config.R.Del(context.Background(), key).Err()
	if err != nil {
		log.Println("删除离线消息失败:", err)
	}
}
