package model

type UserOrder struct {
	BasicModel   `xorm:"extends"` // 包含 Code、Created、Updated 等字段
	UserCode     string  `xorm:"notnull comment('下单用户标识')" json:"user_code"`
	UserName     string  `xorm:"notnull comment('用户名')" json:"user_name"`
	TableNo      string  `xorm:"varchar(20) notnull comment('桌号')" json:"table_no"` // 必填：桌号
	Customer     string  `xorm:"varchar(100) comment('顾客称呼')" json:"customer"`    // 可选：顾客名字或昵称
	TotalPrice   float64 `xorm:"decimal(10,2) notnull comment('订单总价')" json:"total_price"`
	CouponAmount float64 `xorm:"decimal(10,2) notnull comment('折扣后总价')" json:"coupon_amount"`
	CouponCode   string  `xorm:"notnull comment('折扣券价格')" json:"coupon_code"`
	Status       int     `xorm:"default 1 comment('订单状态 1待支付 2制作中 3已完成 4已取消')" json:"status"`
	Remark       string  `xorm:"varchar(255) comment('备注信息')" json:"remark"`
	Version      int     `xorm:"default 1 comment('乐观锁版本号')" json:"version"`
}

type OrderDetail struct {
	BasicModel  `xorm:"extends"`                                                                     // 包含 Code、Created、Updated 等字段
	OrderCode   string  `xorm:"varchar(70) notnull index comment('订单编号外键')" json:"order_code"` // 外键关联主订单表
	ProductCode string  `xorm:"varchar(70) notnull comment('商品编码')" json:"product_code"`         // 商品编码
	ProductName string  `xorm:"varchar(255) notnull comment('商品名称')" json:"product_name"`        // 商品名称（冗余）
	Quantity    int     `xorm:"notnull default 1 comment('商品数量')" json:"quantity"`               // 购买数量
	Price       float64 `xorm:"decimal(10,2) notnull comment('单价')" json:"price"`                  // 单价
	Picture     string  `xorm:"varchar(255) comment('商品图片URL')" json:"picture"`                  // 可选图片地址
}
