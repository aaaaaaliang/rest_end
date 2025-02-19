package model

// RolePermission 角色-权限 关联表
type RolePermission struct {
	BasicModel     `xorm:"extends"`
	RoleCode       string `json:"role_code" xorm:"varchar(70) notnull index comment('角色唯一标识')"`
	PermissionCode string `json:"permission_code" xorm:"varchar(100) notnull index comment('权限唯一标识')"`
}
