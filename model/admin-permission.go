package model

type APIPermission struct {
	BasicModel  `xorm:"extends"`
	ParentCode  *string `json:"parent_code" xorm:"varchar(70) null index"` // 允许为空，表示顶级权限
	Name        string  `json:"name" xorm:"varchar(100)"`
	Method      *string `json:"method" xorm:"varchar(10) null"` // 可为空，顶级分类不需要 HTTP 方法
	Path        *string `json:"path" xorm:"varchar(255) null"`  // 可为空，顶级分类不需要 API 路径
	Description string  `json:"description" xorm:"varchar(255)"`
	Public      int     `json:"public" xorm:"tinyint(1) default(0)"` // 默认值 0，表示公开
}

func (a *APIPermission) TableName() string {
	return "api_permission"
}
