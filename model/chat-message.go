package model

// ChatMessage 聊天消息
type ChatMessage struct {
	BasicModel  `xorm:"extends"`
	SessionCode string `xorm:"varchar(70) notnull comment('会话code')" json:"session_code"` // 用于关联 ChatSession
	SenderCode  string `xorm:"varchar(70) notnull comment('发送方code')" json:"sender_code"` // 顾客或客服的 code
	SenderType  string `xorm:"varchar(20) notnull comment('发送方类型：customer/agent')" json:"sender_type"`
	Content     string `xorm:"text notnull comment('聊天内容')" json:"content"`
}
