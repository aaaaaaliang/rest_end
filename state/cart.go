package state

import "time"

// 购物车Redis 产品
const (
	RedisStockKey  = "stock:%v"       // Redis 中的库存键格式
	RedisPriceKey  = "price:%v"       // Redis 中的价格键格式
	RedisCartLock  = "cart:lock:%s"   // Redis 分布式锁的键格式
	LockExpireTime = 10 * time.Second // 分布式锁的过期时间，防止死锁
)
