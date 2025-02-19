package model

type Products struct {
	BasicModel   `xorm:"extends"` // 基础字段
	ProductsName string           `xorm:"varchar(255) notnull comment('产品名称')" json:"products_name"`
	CategoryCode string           `xorm:"varchar(70) notnull comment('分类唯一标识符')" json:"category_code"`
	Price        float64          `xorm:"decimal(10,2) notnull comment('产品价格')" json:"price"`
	Count        int64            `xorm:"default 0 comment('产品库存数量')" json:"count"`
	Describe     string           `xorm:"text comment('产品描述')" json:"describe"`
	Picture      Annex            `xorm:"json comment('产品图片信息')" json:"picture"`
	Main         int              `xorm:"default 0 comment('特色产品')" json:"main"`
}

func (p *Products) TableName() string {
	return "products"
}
