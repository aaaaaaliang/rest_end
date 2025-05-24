package model

type UserCoupon struct {
	BasicModel   `xorm:"extends"`
	UserCode     string `xorm:"varchar(70) index user_code" json:"user_code"`
	TemplateCode string `xorm:"varchar(70) index template_code" json:"template_code"`
	Status       int    `xorm:"default 0" json:"status"` // 0=未使用, 1=已使用, 2=过期
	ReceiveTime  int64  `xorm:"int" json:"receive_time"` // 发券时间
	ExpireTime   int64  `xorm:"int" json:"expire_time"`  // 过期时间
	UseTime      int64  `xorm:"int" json:"use_time"`     // 使用时间
}
