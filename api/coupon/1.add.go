package coupon

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
	"time"
)

// 创建券模板并初始化库存
func createCouponTemplate(c *gin.Context) {
	type CreateCouponTemplateReq struct {
		Name      string  `json:"name" binding:"required"`
		Type      string  `json:"type" binding:"required"`
		Quota     float64 `json:"quota" binding:"required"`
		MinAmount float64 `json:"min_amount"`
		Total     int     `json:"total" binding:"required"`
		GrantType string  `json:"grant_type" binding:"required"`
		ValidDays int     `json:"valid_days"`
		StartTime int64   `json:"start_time"`
		EndTime   int64   `json:"end_time"`
	}

	var req CreateCouponTemplateReq
	if !utils.ValidationJson(c, &req) {
		return
	}

	creator, ok := utils.GetUserCode(c)
	if !ok {
		response.Success(c, response.ServerError)
		return
	}

	tpl := model.CouponTemplate{
		Name:      req.Name,
		Type:      req.Type,
		Quota:     req.Quota,
		MinAmount: req.MinAmount,
		Total:     req.Total,
		GrantType: req.GrantType,
		ValidDays: req.ValidDays,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    1,
	}
	tpl.Creator = creator

	if _, err := config.DB.Insert(&tpl); err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 初始化 Redis 库存：仅限秒杀券
	if tpl.GrantType == "seckill" {
		stockKey := fmt.Sprintf("coupon:stock:%s", tpl.Code)
		err := config.R.Set(ctx, stockKey, tpl.Total, 0).Err()
		if err != nil {
			log.Printf("❌ Redis 初始化库存失败: %v", err)
		} else {
			log.Printf("✅ Redis 初始化库存成功: %s = %d", stockKey, tpl.Total)
		}
	}

	log.Printf("✅ 创建优惠券模板成功: %+v", tpl)
	response.Success(c, response.SuccessCode)
}

// 手动领取优惠券（manual 类型）
func receiveCoupon(c *gin.Context) {
	type ReceiveCouponReq struct {
		TemplateCode string `json:"template_code" binding:"required"`
	}
	var req ReceiveCouponReq
	if !utils.ValidationJson(c, &req) {
		return
	}

	userCode := utils.GetUser(c)

	// 查询券模板
	var tpl model.CouponTemplate
	has, err := config.DB.Where("code = ? AND grant_type = ? AND status = 1", req.TemplateCode, "manual").Get(&tpl)
	if err != nil || !has {
		response.Success(c, response.BadRequest, errors.New("券模板不存在或不可领取"))
		return
	}

	// 检查是否已领取
	exist, _ := config.DB.Where("user_code = ? AND template_code = ?", userCode, tpl.Code).Exist(new(model.UserCoupon))
	if exist {
		response.Success(c, response.ServerError, errors.New("不可重复领取"))
		return
	}

	// 检查库存
	if tpl.Received >= tpl.Total {
		response.Success(c, response.ServerError, errors.New("优惠券已抢光"))
		return
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
		response.Success(c, response.ServerError, err)
		return
	}

	// 同步扣减库存：received +1，total -1
	res, err := session.Exec("UPDATE coupon_template SET received = received + 1, total = total - 1 WHERE code = ? AND received < total", tpl.Code)
	af, _ := res.RowsAffected()
	if err != nil || af == 0 {
		_ = session.Rollback()
		response.Success(c, response.ServerError, errors.New("库存不足"))
		return
	}

	_ = session.Commit()
	response.Success(c, response.SuccessCode)
}
