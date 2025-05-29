package model

type AIModelConfig struct {
	BasicModel     `xorm:"extends"`
	ModelName      string `json:"model_name" xorm:"varchar(100) notnull comment('模型名称')"` // 如：阿亮餐厅助手
	PromptIntro    string `json:"prompt_intro" xorm:"text comment('提示词')"`                 // 系统提示词（prompt）
	UserLabel      string `json:"user_label" xorm:"varchar(50) default '用户'"`               // 如“用户”
	AssistantLabel string `json:"assistant_label" xorm:"varchar(50) default 'AI客服'"`        // 如“AI客服”
	MaxHistory     int    `json:"max_history" xorm:"default 6 comment('最多上下文轮数')"`     // 控制上下文窗口
}
