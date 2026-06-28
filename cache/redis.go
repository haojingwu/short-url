package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

// InitRedis初始化Redis连接
func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:         "localhost:6379", //Redis地址
		Password:     "",               //无密码
		DB:           0,                //使用默认DB
		PoolSize:     10,               //连接池大小
		MinIdleConns: 5,                //最小空闲连接数
		MaxRetries:   3,                //最大重试次数
	})

	//测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatal("❌ Redis 连接失败", err)
	}
	fmt.Println("✅ Redis 连接成功")
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
