package model

// AIChatHistory 聊天记录表
type AIChatHistory struct {
	BasicModel  `xorm:"extends"`
	UserCode    string `json:"user_code" xorm:"varchar(70) index notnull comment('用户ID')"`
	SessionCode string `json:"session_code" xorm:"varchar(70) index notnull comment('会话ID')"`
	Role        string `json:"role" xorm:"varchar(20) comment('角色:user 或 assistant')"`
	Content     string `json:"content" xorm:"text comment('对话内容')"`
}
