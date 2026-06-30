package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

// InitRedis初始化Redis连接
func InitRedis() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		fmt.Sscanf(dbStr, "%d", &db)
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	RDB = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatal("❌ Redis 连接失败:", err)
	}

	fmt.Printf("✅ Redis 连接成功: %s\n", addr)
}

// GetCache 从Redis获取缓存
func GetCache(ctx context.Context, key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// SetCache 写入Redis缓存(带过期时间)
func SetCache(ctx context.Context, key string, value string, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

// DeleteCache 删除Redis缓存
func DeleteCache(ctx context.Context, key string) error {
	return RDB.Del(ctx, key).Err()
}
