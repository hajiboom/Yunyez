// Package redis Redis Client
// 提供全局唯一的 Redis 客户端实例
package redis

import (
	"context"
	"time"
	"sync"

	config "yunyez/internal/common/config"

	"github.com/redis/go-redis/v9"
	logger "yunyez/internal/pkg/logger"
)

var (
	RedisClient *Client
	once sync.Once
)

// Client Redis 客户端结构体
type Client struct {
	*redis.Client
}

// NewClient create redis client
func NewClient() (*Client, error) {
	redisAddr := config.GetString("redis.addr")
	redisPassword := config.GetString("redis.password")
	redisDB := config.GetInt("redis.db")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Error(ctx, "Failed to connect to Redis", map[string]interface{}{
			"addr": redisAddr,
			"db":   redisDB,
		})
		return nil, err
	}

	return &Client{rdb}, nil
}

// GetRedisClient Get Redis Client
func GetRedisClient() (*Client, error) {
	once.Do(func() {
		client, err := NewClient()
		if err != nil {
			panic(err)
		}
		RedisClient = client
	})
	return RedisClient, nil
}