package model

// ChatSession 顾客与客服的一对一会话
type ChatSession struct {
	BasicModel   `xorm:"extends"`
	CustomerCode string `xorm:"varchar(70) notnull comment('顾客code')" json:"customer_code"`
	AgentCode    string `xorm:"varchar(70) notnull comment('客服code')" json:"agent_code"`
	Status       string `xorm:"varchar(20) default 'active' comment('会话状态：active/ended')" json:"status"`
}
