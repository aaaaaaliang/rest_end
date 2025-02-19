package model

type Banner struct {
	BasicModel `xorm:"extends"`
	Image      string `json:"image" xorm:"varchar(255) notnull"`                    // 图片地址
	Title      string `json:"title" xorm:"varchar(255)"`                            // 轮播图标题
	Sort       int    `json:"sort" xorm:"int default(0)"`                           // 排序
	Category   int    `json:"category" xorm:"int default(0) comment('0 轮播图  1菜品')"` // 图片分类
}
