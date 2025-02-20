package model

// UserRole 用户-角色 关联表
type UserRole struct {
	BasicModel `xorm:"extends"`
	UserCode   string `xorm:"'user_code' index"` // 使用小写字段名
	RoleCode   string `xorm:"'role_code' index"` // 使用小写字段名
}
