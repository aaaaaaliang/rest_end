package coupon

import (
	"github.com/gin-gonic/gin"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

// listTemplates 分页查询优惠券模板
func listTemplates(c *gin.Context) {
	type ListTemplateReq struct {
		Index int `form:"index" binding:"required,min=1"`
		Size  int `form:"size"  binding:"required,min=1,max=100"`
	}

	var req ListTemplateReq
	if !utils.ValidationQuery(c, &req) {
		return
	}

	var templates []model.CouponTemplate
	total, err := config.DB.Table(new(model.CouponTemplate)).
		Limit(req.Size, (req.Index-1)*req.Size).
		FindAndCount(&templates)

	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	response.SuccessWithTotal(c, response.SuccessCode, templates, int(total))
}

// /coupon/list 查询我的所有券（标记是否拥有）
func listAllCoupons(c *gin.Context) {
	userCode := utils.GetUser(c)

	// 1. 查所有券模板（不限制类型）
	var templates []model.CouponTemplate
	err := config.DB.Where("status = 1").Find(&templates)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	// 2. 查用户已拥有的券模板 code
	var userCoupons []model.UserCoupon
	_ = config.DB.Where("user_code = ?", userCode).Find(&userCoupons)
	owned := map[string]bool{}
	for _, c := range userCoupons {
		owned[c.TemplateCode] = true
	}

	// 3. 构造响应
	type Res struct {
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		Type      string  `json:"type"`
		Quota     float64 `json:"quota"`
		GrantType string  `json:"grant_type"`
		Owned     bool    `json:"owned"` // 是否已领取
	}
	var result []Res
	for _, tpl := range templates {
		// login 类型仅当 owned 为 true 才显示（避免未自动发放的尴尬）
		if tpl.GrantType == "login" && !owned[tpl.Code] {
			continue
		}
		result = append(result, Res{
			Code:      tpl.Code,
			Name:      tpl.Name,
			Type:      tpl.Type,
			Quota:     tpl.Quota,
			GrantType: tpl.GrantType,
			Owned:     owned[tpl.Code],
		})
	}

	response.SuccessWithData(c, response.SuccessCode, result)
}

// 查询当前用户的所有优惠券（包含状态）
func getUserCoupons(c *gin.Context) {
	userCode := utils.GetUser(c)

	var coupons []model.UserCoupon
	err := config.DB.Where("user_code = ?", userCode).OrderBy("expire_time ASC").Find(&coupons)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	type Res struct {
		Code           string  `json:"code"`
		TemplateCode   string  `json:"template_code"`
		Status         int     `json:"status"`
		ReceiveTime    int64   `json:"receive_time"`
		ExpireTime     int64   `json:"expire_time"`
		MinAmount      float64 `json:"min_amount"`
		Name           string  `json:"name"`
		Type           string  `json:"type"`
		Quota          float64 `json:"quota"`
		DiscountAmount float64 `json:"discount_amount"`
	}

	simulateDiscount := func(tplType string, quota float64, minAmount float64, testPrice float64) float64 {
		if testPrice < minAmount {
			return 0
		}
		switch tplType {
		case "full", "cash":
			if quota > testPrice {
				return testPrice
			}
			return quota
		case "discount":
			discount := testPrice * (1 - quota)
			return discount
		default:
			return 0
		}
	}

	var res []Res
	for _, uc := range coupons {
		var tpl model.CouponTemplate
		has, err := config.DB.Where("code = ?", uc.TemplateCode).Get(&tpl)
		if err != nil || !has {
			continue
		}
		discount := simulateDiscount(tpl.Type, tpl.Quota, tpl.MinAmount, 100)
		res = append(res, Res{
			Code:           uc.Code,
			TemplateCode:   uc.TemplateCode,
			Status:         uc.Status,
			ReceiveTime:    uc.ReceiveTime,
			ExpireTime:     uc.ExpireTime,
			MinAmount:      tpl.MinAmount,
			Name:           tpl.Name,
			Type:           tpl.Type,
			Quota:          tpl.Quota,
			DiscountAmount: discount,
		})
	}

	response.SuccessWithData(c, response.SuccessCode, res)
}
