package model

type Category struct {
	BasicModel    `xorm:"extends"`
	CategoryName  string     `xorm:"varchar(255) notnull" comment:"分类名称"`
	ParentCode    *string    `xorm:"varchar(70) null index" comment:"祖先code"` // 父级分类唯一标识
	SubCategories []Category `json:"sub_categories" xorm:"-"`                 // 添加一个字段来存储子分类数据
	Sort          int        `xorm:"int default(0)" comment:"排序"`
}

func (c *Category) TableName() string {
	return "category"
}
