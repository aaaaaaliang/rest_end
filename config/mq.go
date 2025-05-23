package config

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type MqConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Url      string `mapstructure:"url"`
}

var MQConn *amqp.Connection

func InitMQ() {
	dsn := fmt.Sprintf("amqp://%v:%v@%v/", G.MQ.Username, G.MQ.Password, G.MQ.Url)
	conn, err := amqp.Dial(dsn)
	if err != nil {
		log.Fatalf("❌ RabbitMQ 连接失败: %v", err)
	}
	log.Println("✅ RabbitMQ 连接成功")
	MQConn = conn

	// 初始化 RabbitMQ 延时队列和死信队列
	ch, err := GetRabbitMQChannel()
	if err != nil {
		log.Fatalf("获取 RabbitMQ Channel 失败: %v", err)
	}
	// 初始化延时队列
	if err := setupDelayQueue(ch); err != nil {
		log.Fatalf("初始化延时队列失败: %v", err)
	}
	// 初始化死信队列用于超时处理
	if err := setupTimeoutQueue(ch); err != nil {
		log.Fatalf("初始化死信队列失败: %v", err)
	}
	ch.Close()
}

// GetRabbitMQChannel ✅ 获取 RabbitMQ 通道（每个 goroutine 需要单独获取）
func GetRabbitMQChannel() (*amqp.Channel, error) {
	if MQConn == nil {
		return nil, fmt.Errorf("❌ RabbitMQ 连接未初始化")
	}
	return MQConn.Channel() // 每次调用返回一个新的 Channel
}

// 初始化延时队列及其死信机制
func setupDelayQueue(ch *amqp.Channel) error {

	// 1️⃣ 声明一个死信交换机（用于接收延时队列中过期的消息）
	//    类型使用 direct，支持按 routing key 精确投递
	err := ch.ExchangeDeclare(
		"order_timeout_exchange", // 交换机名称
		"direct",                 // 交换机类型（直连）
		true,                     // 持久化
		false,                    // 无绑定队列时不自动删除
		false,                    // 非内部交换机
		false,                    // 等待服务器确认
		nil,                      // 无额外参数
	)
	if err != nil {
		return fmt.Errorf("声明死信交换机失败: %v", err)
	}

	// 2️⃣ 声明延时队列（用于临时存放新订单消息）
	//    设置消息 TTL（存活时间），超过时间则自动过期
	//    并指定死信交换机与 routing key，用于处理过期后的消息
	args := amqp.Table{
		"x-message-ttl":             int32(900000),            // 消息 TTL：15分钟（单位毫秒）
		"x-dead-letter-exchange":    "order_timeout_exchange", // 过期后将消息发送到哪个交换机
		"x-dead-letter-routing-key": "order_timeout",          // 转发时使用的路由键
	}
	_, err = ch.QueueDeclare(
		"order_delay_queue", // 队列名称
		true,                // 持久化
		false,               // 不自动删除
		false,               // 不独占
		false,               // 等待服务器响应
		args,                // 附加参数（TTL + 死信配置）
	)
	if err != nil {
		return fmt.Errorf("声明延时队列失败: %v", err)
	}
	return nil
}

// 初始化死信队列
func setupTimeoutQueue(ch *amqp.Channel) error {
	// 声明一个死信队列，用于接收过期消息
	_, err := ch.QueueDeclare("order_timeout_queue", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("声明死信队列失败: %v", err)
	}
	// 绑定队列到死信交换机
	err = ch.QueueBind("order_timeout_queue", "order_timeout", "order_timeout_exchange", false, nil)
	if err != nil {
		return fmt.Errorf("绑定死信队列失败: %v", err)
	}
	return nil
}
