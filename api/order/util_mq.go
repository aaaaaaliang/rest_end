package order

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"rest/config"
	"rest/model"
)

func publishMessage(queueName string, message []byte) error {
	ch, err := config.GetRabbitMQChannel()
	if err != nil {
		return err
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Fatalf("Failed to close RabbitMQ channel")
		}
	}(ch) // ✅ 用完就关闭 Channel，防止并发冲突

	// 声明队列
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 发送消息  direct交换机
	err = ch.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        message,
	})
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Println("📤 消息已发送:", queueName)
	return nil
}

func publishDelayOrder(order *model.UserOrder) error {
	ch, err := config.GetRabbitMQChannel()
	if err != nil {
		return err
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Fatalf("Failed to close RabbitMQ channel %v", err)
		}
	}(ch)

	message, err := json.Marshal(order)
	if err != nil {
		return err
	}
	err = ch.Publish("", "order_delay_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        message,
	})
	if err != nil {
		return fmt.Errorf("发布延时消息失败: %v", err)
	}
	log.Printf("延时消息已发布，订单号为 %v 将在 15 分钟后进行超时检测", order.Code)
	return nil
}

func ConsumeOrderMessages() {
	ch, err := config.GetRabbitMQChannel()
	if err != nil {
		log.Fatalf("❌ 获取 RabbitMQ Channel 失败: %v", err)
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Fatalf("Failed to close RabbitMQ channel")
		}
	}(ch)

	// 声明队列（如果发送时已经声明过可以省略）
	_, err = ch.QueueDeclare("order_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ 声明队列失败: %v", err)
	}

	// 使用手动确认模式，确保消息正确处理后再确认
	msg, err := ch.Consume("order_queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ 监听队列失败: %v", err)
	}

	log.Println("🐰 [*] 开始监听订单队列")

	for msg := range msg {
		log.Println("📥 收到订单消息:", string(msg.Body))
		// 解析订单数据
		var order model.UserOrder
		if err := json.Unmarshal(msg.Body, &order); err != nil {
			log.Printf("解析订单消息失败: %v", err)
			_ = msg.Nack(false, false)
			continue
		}
		// 异步处理：例如存入 Elasticsearch
		if err := saveOrderToES(&order); err != nil {
			log.Printf("存入 Elasticsearch 失败: %v", err)
			// 这里可以设置重试或者放入死信队列
			_ = msg.Nack(false, true)
			continue
		}
		_ = msg.Ack(false) // 处理成功确认消息
	}
}

func ConsumeTimeoutMessages() {
	ch, err := config.GetRabbitMQChannel()
	if err != nil {
		log.Fatalf("❌ 获取 RabbitMQ Channel 失败: %v", err)
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Fatalf("死信队列关闭失败 %v", err)
		}
	}(ch)

	msg, err := ch.Consume("order_timeout_queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ 监听死信队列失败: %v", err)
	}

	log.Println("🐰 [*] 开始监听订单超时队列")
	for msg := range msg {
		log.Println("📥 收到订单超时消息:", string(msg.Body))
		var order model.UserOrder
		if err := json.Unmarshal(msg.Body, &order); err != nil {
			log.Printf("解析订单消息失败: %v", err)
			_ = msg.Nack(false, false)
			continue
		}

		// 检查订单状态是否为待支付（1）
		if order.Status == 1 {
			// 1. 更新订单状态为取消
			if err := updateOrderStatus(order.Code, 4); err != nil {
				log.Printf("更新订单状态失败: %v", err)
				_ = msg.Nack(false, true)
				continue
			}

			// 2. 查询订单明细
			var details []model.OrderDetail
			if err := config.DB.Where("order_code = ?", order.Code).Find(&details); err != nil {
				log.Printf("查询订单明细失败: %v", err)
				_ = msg.Nack(false, true)
				continue
			}

			// 3. 回补库存
			for _, d := range details {
				_, err := config.DB.Exec(
					"UPDATE products SET count = count + ? WHERE code = ?", d.Quantity, d.ProductCode,
				)
				if err != nil {
					log.Printf("商品 %s 库存回补失败: %v", d.ProductCode, err)
					_ = msg.Nack(false, true)
					continue
				}
			}

			// 4. 更新 ES 状态
			order.Status = 4 // 已取消
			err := updateOrderInES(&order)
			if err != nil {
				log.Printf("更新 Elasticsearch 失败: %v", err)
				_ = msg.Nack(false, true)
				continue
			}

			log.Printf("✅ 订单 %s 已超时取消，库存已回补", order.Code)
		}

		// 消息处理成功
		_ = msg.Ack(false)
	}
}

func updateOrderStatus(orderCode string, newStatus int) error {
	_, err := config.DB.Table(model.UserOrder{}).Where("code = ?", orderCode).Update(map[string]interface{}{
		"status": newStatus,
	})
	return err
}
