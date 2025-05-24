package coupon

import "github.com/gin-gonic/gin"

func RegisterCouponTemplateRoutes(group *gin.RouterGroup) {
	group.POST("/coupon/template", createCouponTemplate)
	group.GET("/coupon/template", listTemplates)
	group.PUT("/coupon/template/:code", updateTemplate)
	group.PUT("/coupon/template/status/:code", changeTemplateStatus)
	group.DELETE("/coupon/template/:code", deleteTemplate) // 删除（逻辑）
	group.POST("/coupon/receive", receiveCoupon)           // 领券
	group.GET("/coupon/list", listAllCoupons)              // 所有券 + 拥有标记
	group.POST("/coupon/seckill/receive", seckillCoupon)   // 秒杀券
	group.GET("/coupon/seckill", getUserCoupons)

}
