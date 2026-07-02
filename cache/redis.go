package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"short-url/config"
)

var RDB *redis.Client

func InitRedis() {
	cfg := config.Cfg

	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatal("❌ Redis 连接失败:", err)
	}

	log.Printf("✅ Redis 连接成功: %s\n", cfg.Redis.Addr)
}

// GetCache 从 Redis 获取缓存
func GetCache(ctx context.Context, key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// SetCache 写入 Redis 缓存
func SetCache(ctx context.Context, key string, value string, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

// DeleteCache 删除 Redis 缓存
func DeleteCache(ctx context.Context, key string) error {
	return RDB.Del(ctx, key).Err()
}
