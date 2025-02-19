package model

type Role struct {
	BasicModel  `xorm:"extends"`
	Name        string `json:"name" xorm:"varchar(100) unique notnull comment('角色名称')"`
	Description string `json:"description" xorm:"varchar(255) comment('角色描述')"`
}
