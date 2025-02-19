package model

type SalaryRecord struct {
	BasicModel  `xorm:"extends"`
	UserCode    string  `xorm:"index notnull comment('用户唯一标识')" json:"user_code"`
	BaseSalary  float64 `xorm:"decimal(10,2) notnull comment('基本工资')" json:"base_salary" `
	Bonus       float64 `xorm:"decimal(10,2) notnull default 0.00 comment('奖金')" json:"bonus"`
	Deduction   float64 `xorm:"decimal(10,2) notnull default 0.00 comment('扣款')" json:"deduction" `
	TotalSalary float64 `xorm:"decimal(10,2) notnull comment('最终工资')" json:"total_salary" `
	PayDate     int64   `xorm:"bigint notnull comment('发放日期')" json:"pay_date"`
}
