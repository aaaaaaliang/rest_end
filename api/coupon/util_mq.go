package coupon

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"rest/config"
	"rest/model"
	"time"
)

// StartCouponConsumer 启动消费 coupon_timeout_queue 中的发券消息
func StartCouponConsumer() {
	go func() {
		ch, err := config.GetRabbitMQChannel()
		if err != nil {
			log.Fatalf("❌ 获取 RabbitMQ channel 失败: %v", err)
		}
		defer ch.Close()

		msgs, err := ch.Consume(
			"coupon_timeout_queue", // 队列名
			"",                     // consumer
			true,                   // auto-ack
			false, false, false, nil,
		)
		if err != nil {
			log.Fatalf("❌ 订阅 coupon_timeout_queue 失败: %v", err)
		}

		log.Println("🚀 正在监听 coupon_timeout_queue 队列...")
		for msg := range msgs {
			decoded, err := base64.StdEncoding.DecodeString(string(msg.Body))
			if err != nil {
				log.Printf("❌ Base64 解码失败: %v", err)
				continue
			}

			var data map[string]string
			err = json.Unmarshal(decoded, &data)
			if err != nil {
				log.Printf("❌ JSON 解析失败: %v", err)
				continue
			}

			userCode := data["user_code"]
			templateCode := data["template_code"]

			// 查找模板有效性
			var tpl model.CouponTemplate
			has, err := config.DB.Where("code = ? AND status = 1", templateCode).Get(&tpl)
			if err != nil || !has {
				log.Printf("❌ 模板无效: %v", templateCode)
				continue
			}

			// 检查是否重复发券
			exist, _ := config.DB.Where("user_code = ? AND template_code = ?", userCode, templateCode).Exist(&model.UserCoupon{})
			if exist {
				log.Printf("⚠️ 用户[%s] 已领取过模板[%s]", userCode, templateCode)
				continue
			}

			now := time.Now().Unix()
			expire := now + int64(tpl.ValidDays*86400)

			uc := model.UserCoupon{
				UserCode:     userCode,
				TemplateCode: templateCode,
				Status:       0,
				ReceiveTime:  now,
				ExpireTime:   expire,
			}

			session := config.DB.NewSession()
			defer session.Close()
			_ = session.Begin()

			if _, err := session.Insert(&uc); err != nil {
				_ = session.Rollback()
				log.Printf("❌ 插入用户券失败: %v", err)
				continue
			}

			// 同步更新数据库中 coupon_template 的库存（received +1，total -1）
			res, err := session.Exec("UPDATE coupon_template SET received = received + 1, total = total - 1 WHERE code = ? AND received < total", templateCode)
			af, _ := res.RowsAffected()
			if err != nil || af == 0 {
				_ = session.Rollback()
				log.Printf("❌ 数据库扣减库存失败或库存不足: %v", err)
				continue
			}
			_ = session.Commit()
			log.Printf("🎁 发券成功: 用户[%s] 领取模板[%s]，库存同步扣减", userCode, templateCode)
		}
	}()
}
