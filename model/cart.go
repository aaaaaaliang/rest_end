package model

import (
	"fmt"
	"github.com/xormplus/xorm"
)

// UserCart 购物车模型
type UserCart struct {
	BasicModel  `xorm:"extends"`
	UserCode    string  `xorm:"user_code comment('用户code') index(user_product_index)" json:"user_code" binding:"required"`
	ProductCode string  `xorm:"varchar(255) comment('产品编号') index(user_product_index)" json:"product_code" binding:"required"`
	ProductNum  int     `xorm:"default 0 comment('产品数量')" json:"product_num" binding:"required,min=1"`
	TotalPrice  float64 `xorm:"decimal(10,2) comment('总价格')" json:"total_price" binding:"required,gt=0"`
	IsOrdered   bool    `xorm:"default false comment('是否已下单')" json:"is_ordered"` // 新增字段
}

// ExistWithCodeAndUser 判断是否存在和获取数据
func (u *UserCart) ExistWithCodeAndUser(session *xorm.Session) (bool, error) {
	if u.ProductCode == "" || u.UserCode == "" {
		return false, fmt.Errorf("product_code 或 user_code 为空")
	}

	b, err := session.Where("product_code = ? AND user_code = ?", u.ProductCode, u.UserCode).Get(u)
	return b, err
}

// UpdateProductNum 更新产品数量和总价
func (u *UserCart) UpdateProductNum(session *xorm.Session) error {
	if u.ProductCode == "" || u.UserCode == "" {
		return fmt.Errorf("product_code 或 user_code 为空")
	}
	// 更新数据库中的记录
	_, err := session.Where("product_code = ? AND user_code = ?", u.ProductCode, u.UserCode).
		Cols("product_num", "total_price").
		Update(u)
	if err != nil {
		return fmt.Errorf("更新 product_num 失败: %w", err)
	}

	return nil
}
