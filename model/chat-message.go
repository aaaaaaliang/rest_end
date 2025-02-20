package model

type ChatMessage struct {
	Id              int64  `json:"id" xorm:"pk autoincr"`
	FromUser        string `json:"from_user"`
	ToUser          string `json:"to_user,omitempty"`
	Content         string `json:"content"`
	Timestamp       int64  `json:"timestamp"`
	Type            string `json:"type"`
	IsHandled       bool   `json:"is_handled"`
	Read            bool   `json:"read"`
	SupportUserCode string `json:"support_user_code,omitempty"`
	Status          string `json:"status,omitempty"`
	Role            string `json:"role,omitempty"` // 可选字段，用于存储角色
}
