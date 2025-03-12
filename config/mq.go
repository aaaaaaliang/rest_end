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

// 初始化延时队列
func setupDelayQueue(ch *amqp.Channel) error {
	// 声明死信交换机（比如 order_timeout_exchange）
	err := ch.ExchangeDeclare("order_timeout_exchange", "direct", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("声明死信交换机失败: %v", err)
	}

	// 声明延时队列，设置 TTL 为 30 分钟（1800000 毫秒），并绑定死信交换机
	args := amqp.Table{
		"x-message-ttl":             int32(900000), // 15分钟过期
		"x-dead-letter-exchange":    "order_timeout_exchange",
		"x-dead-letter-routing-key": "order_timeout",
	}
	_, err = ch.QueueDeclare("order_delay_queue", true, false, false, false, args)
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
