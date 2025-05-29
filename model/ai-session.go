package model

type AIChatSession struct {
	BasicModel   `xorm:"extends"`
	UserCode     string `json:"user_code" xorm:"varchar(70) notnull index comment('用户标识')"`
	SessionTitle string `json:"session_title" xorm:"varchar(255) comment('会话标题，可生成')"`
	LastMessage  string `json:"last_message" xorm:"text comment('最后一条消息内容')"`
}
