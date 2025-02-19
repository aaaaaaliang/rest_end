package model

// UserRole 用户-角色 关联表
type UserRole struct {
	BasicModel `xorm:"extends"`
	UserCode   string `xorm:"index"` // 对应 Users 表的主键ID
	RoleCode   string `xorm:"index"` // 对应 Role 表的主键ID
}
