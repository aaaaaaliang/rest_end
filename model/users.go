package model

type Users struct {
	BasicModel `xorm:"extends"`
	Username   string  `xorm:"varchar(255)" json:"username"`
	Password   string  `xorm:"varchar(255)" json:"password"`
	Email      string  `xorm:"varchar(255) comment('邮箱')" json:"email"`
	Nickname   string  `xorm:"varchar(255) comment('昵称')" json:"nickname"`
	Gender     string  `xorm:"varchar(255) comment('性别')" json:"gender"`
	RealName   string  `xorm:"varchar(255) comment('姓名')" json:"real_name"`
	Phone      string  `xorm:"varchar(255) comment('电话')" json:"phone"`
	BaseSalary float64 `xorm:"decimal(10,2) notnull default 0.00 comment('基本工资')" json:"base_salary"`
	IsEmployee bool    `xorm:"tinyint(1) notnull default 0 comment('是否为员工')" json:"is_employee"`
	//Roles      []Role `xorm:"-"`
}
