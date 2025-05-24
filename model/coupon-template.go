package model

type CouponTemplate struct {
	BasicModel `xorm:"extends"`
	Name       string  `xorm:"varchar(100) notnull comment('券名称')" json:"name" binding:"required"`
	Type       string  `xorm:"varchar(20) notnull comment('券类型: full减免 / discount折扣 / cash现金')" json:"type" binding:"required"`
	Quota      float64 `xorm:"decimal(10,2) notnull comment('满减金额/折扣比例')" json:"quota" binding:"required"`
	MinAmount  float64 `xorm:"decimal(10,2) default(0) comment('使用门槛')" json:"min_amount"`
	Total      int     `xorm:"notnull comment('发放总量')" json:"total" binding:"required"`
	Received   int     `xorm:"default(0) comment('已领取数量')" json:"received"`
	GrantType  string  `xorm:"varchar(20) notnull comment('发放方式: login/manual/seckill')" json:"grant_type" binding:"required"`
	ValidDays  int     `xorm:"int comment('相对有效期: 领取后N天')" json:"valid_days"`
	StartTime  int64   `xorm:"bigint comment('固定开始时间')" json:"start_time,omitempty"`
	EndTime    int64   `xorm:"bigint comment('固定结束时间')" json:"end_time,omitempty"`
	Status     int     `xorm:"default(1) comment('状态: 0=禁用 1=启用')" json:"status"`
}
