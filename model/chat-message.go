package model

import "time"

// ChatMessage 存储聊天记录
type ChatMessage struct {
	ID        int64     `xorm:"pk autoincr 'id'"`                      // 主键，自增 ID
	FromUser  string    `xorm:"varchar(50) notnull 'from_user'"`       // 发送者 user_code
	ToUser    string    `xorm:"varchar(50) 'to_user'"`                 // 接收者 user_code（为空表示群发）
	Content   string    `xorm:"text notnull 'content'"`                // 消息内容
	Type      string    `xorm:"varchar(20) notnull 'type'"`            // 消息类型（text, image, file）
	Status    string    `xorm:"varchar(20) default 'unread' 'status'"` // 消息状态（unread, read）
	Timestamp time.Time `xorm:"created 'timestamp'"`                   // 发送时间
}
