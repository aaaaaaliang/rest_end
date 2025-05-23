package model

type TableInfo struct {
	BasicModel `xorm:"extends"`
	Location   string `xorm:"varchar(100)" json:"location"` // 位置（可选）
	Seats      int    `xorm:"default 4" json:"seats"`       // 座位数
	Status     int    `xorm:"default 1" json:"status"`      // 状态：1可用 0不可用
	Remark     string `xorm:"varchar(255)" json:"remark"`   // 备注
}
