package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"dscheduler/config"
)

// RDB 全局 Redis 客户端
var RDB *redis.Client

// InitRedis 初始化 Redis 客户端
func InitRedis() error {
	c := config.Global.Redis
	RDB = redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return RDB.Ping(ctx).Err()
}

// TryLock 基于 SET NX PX 的分布式锁，用于调度中心多实例部署时防重复触发
func TryLock(ctx context.Context, key string, expire time.Duration) (bool, error) {
	return RDB.SetNX(ctx, key, "1", expire).Result()
}
