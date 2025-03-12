package config

import "sync"

var once sync.Once

func InitConfig() {
	once.Do(func() {
		InitDB()
		InitRedis()
		InitMQ()

	})
}
