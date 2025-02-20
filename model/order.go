package model

type OrderDetail struct {
	ProductCode string  `json:"product_code"` // 商品ID
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"` // 商品数量
	Price       float64 `json:"price"`    // 商品单价
	Picture     string  `json:"picture"`
}

type UserOrder struct {
	BasicModel  `xorm:"extends"` // 继承基础字段
	UserCode    string           `xorm:"notnull comment('用户ID')" json:"user_code"`                                  // 用户ID
	TotalPrice  float64          `xorm:"notnull comment('订单总金额')" json:"total_price"`                               // 总金额
	Status      int              `xorm:"default 0 comment('订单状态  1已下单 2.制作中 3.已完成 4.已逾期（暂留）5.无法处理')" json:"status"` // 订单状态
	Remark      string           `xorm:"varchar(255) comment('订单备注')" json:"remark"`                                // 备注信息
	OrderDetail []OrderDetail    `xorm:"json comment('订单详细信息，存储为JSON格式')" json:"order_detail"`                      // 自动处理 JSON 数据
	Version     int              `xorm:"default 1 comment('乐观锁版本号')" json:"version"`
}
