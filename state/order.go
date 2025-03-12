package state

// OrderStatus 订单状态枚举
type OrderStatus int

const (
	Ordered             OrderStatus = iota + 1 // 1. 已下单
	OrderInProgress                            // 2. 制作中
	OrderCompleted                             // 3. 已完成
	OrderCanceled                              // 4. 取消订单
	OrderPendingPayment                        // 5. 待支付
)
