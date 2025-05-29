package coupon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"log"
	"rest/config"
	"rest/model"
	"rest/response"
	"rest/utils"
)

var ctx = context.Background()

// updateTemplate 编辑模板
func updateTemplate(c *gin.Context) {
	code := c.Param("code")
	var req model.CouponTemplate
	if !utils.ValidationJson(c, &req) {
		return
	}

	affect, err := config.DB.Where("code = ?", code).Update(&req)
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	if affect == 0 {
		response.Success(c, response.NotFound, errors.New("记录不存在"))
		return
	}
	response.Success(c, response.SuccessCode)
}

// changeTemplateStatus 启用/禁用模板
func changeTemplateStatus(c *gin.Context) {
	code := c.Param("code")
	type Req struct {
		Status *int `json:"status" binding:"required"`
	}
	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}
	if req.Status == nil || (*req.Status != 0 && *req.Status != 1) {
		response.Success(c, response.ServerError, errors.New("status 只能是 0 或 1"))
		return
	}

	affect, err := config.DB.Where("code = ?", code).Cols("status").Update(&model.CouponTemplate{Status: *req.Status})
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}
	if affect == 0 {
		response.Success(c, response.NotFound, errors.New("记录不存在"))
		return
	}
	response.Success(c, response.SuccessCode)
}

// /coupon/seckill/receive 秒杀抢券（Redis + Lua 限流 + MQ）
func seckillCoupon(c *gin.Context) {
	type Req struct {
		TemplateCode string `json:"template_code" binding:"required"`
	}
	var req Req
	if !utils.ValidationJson(c, &req) {
		return
	}
	userCode := utils.GetUser(c)

	stockKey := fmt.Sprintf("coupon:stock:%s", req.TemplateCode)
	userFlagKey := fmt.Sprintf("coupon:seckill:%s:%s", req.TemplateCode, userCode)

	script := redis.NewScript(`
		if redis.call("exists", KEYS[2]) == 1 then
			return -1
		end
		if tonumber(redis.call("get", KEYS[1])) <= 0 then
			return 0
		end
		redis.call("decr", KEYS[1])
		redis.call("set", KEYS[2], 1)
		return 1
	`)

	res, err := script.Run(ctx, config.R, []string{stockKey, userFlagKey}).Int()
	if err != nil {
		response.Success(c, response.ServerError, err)
		return
	}

	switch res {
	case -1:
		response.Success(c, response.ServerError, errors.New("您已抢过该券"))
		return
	case 0:
		response.Success(c, response.ServerError, errors.New("优惠券已被抢光"))
		return
	case 1:
		// 秒杀成功 → 写入 MQ
		body, _ := json.Marshal(map[string]string{
			"user_code":     userCode,
			"template_code": req.TemplateCode,
		})
		encoded := base64.StdEncoding.EncodeToString(body)
		ch, err := config.GetRabbitMQChannel()
		if err != nil {
			response.Success(c, response.ServerError, err)
			return
		}
		defer func(ch *amqp.Channel) {
			err := ch.Close()
			if err != nil {
				log.Printf("ch.Close error: %v", err)
			}
		}(ch)

		err = ch.Publish(
			"order_timeout_exchange", // 指定交换机
			"coupon_timeout",         // coupon 专属 routing key
			false, false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(encoded),
			},
		)
		if err != nil {
			response.Success(c, response.ServerError, err)
			return
		}
		response.Success(c, response.SuccessCode)
		return
	default:
		response.Success(c, response.ServerError, errors.New("未知响应"))
		return
	}
}
