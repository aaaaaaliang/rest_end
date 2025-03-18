package cart

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/xormplus/xorm"
	"log"
	"rest/config"
	"rest/response"
	"rest/state"
)

// acquireLock 获取分布式锁
func acquireLock(ctx context.Context, lockKey string) bool {
	ok, err := config.R.SetNX(ctx, lockKey, "1", state.LockExpireTime).Result()
	if err != nil || !ok {
		log.Printf("Failed to acquire lock for %s, err: %v", lockKey, err)
		return false
	}
	return true
}

// releaseLock 释放分布式锁
func releaseLock(ctx context.Context, lockKey string) {
	if _, err := config.R.Del(ctx, lockKey).Result(); err != nil {
		log.Printf("Failed to release lock for %s, err: %v", lockKey, err)
	}
}

// rollbackTransaction 统一事务回滚逻辑
func rollbackTransaction(session *xorm.Session, c *gin.Context, err error) {
	_ = session.Rollback()
	response.Success(c, response.UpdateFail, err)
}
