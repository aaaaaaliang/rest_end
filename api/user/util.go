package user

import (
	"context"
	"fmt"
	"rest/config"
	"rest/model"
	"time"
)

// grantLoginCoupons 登录发券逻辑（带库存同步）
func grantLoginCoupons(ctx context.Context, userCode string) {
	var templates []model.CouponTemplate

	select {
	case <-ctx.Done():
		fmt.Println("⚠️ grantLoginCoupons 超时取消，查询模板跳过")
		return
	default:
	}

	err := config.DB.Where("grant_type = ? AND status = 1", "login").Find(&templates)
	if err != nil {
		fmt.Printf("❌ 查询登录券失败: %v\n", err)
		return
	}

	for _, tpl := range templates {
		select {
		case <-ctx.Done():
			fmt.Printf("⚠️ grantLoginCoupons 被取消，跳出发券循环\n")
			return
		default:
		}

		exist, _ := config.DB.Where("user_code = ? AND template_code = ?", userCode, tpl.Code).Exist(new(model.UserCoupon))
		if exist {
			continue
		}

		if tpl.Received >= tpl.Total {
			continue
		}

		now := time.Now().Unix()
		expire := now + int64(tpl.ValidDays*86400)

		userCoupon := model.UserCoupon{
			UserCode:     userCode,
			TemplateCode: tpl.Code,
			Status:       0,
			ReceiveTime:  now,
			ExpireTime:   expire,
		}

		session := config.DB.NewSession()
		defer session.Close()
		_ = session.Begin()

		if _, err := session.Insert(&userCoupon); err != nil {
			_ = session.Rollback()
			continue
		}

		res, err := session.Exec("UPDATE coupon_template SET received = received + 1, total = total - 1 WHERE code = ? AND received < total", tpl.Code)
		af, _ := res.RowsAffected()
		if err != nil || af == 0 {
			_ = session.Rollback()
			continue
		}

		_ = session.Commit()
		fmt.Printf("🎁 发放登录券成功: 模板[%s] → 用户[%s]\n", tpl.Code, userCode)
	}
}
