package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist Token 黑名单管理器
type TokenBlacklist struct {
	redisClient *redis.Client
	config      RedisConfig
}

// NewTokenBlacklist 创建 Token 黑名单管理器
func NewTokenBlacklist(client *redis.Client, config RedisConfig) *TokenBlacklist {
	return &TokenBlacklist{
		redisClient: client,
		config:      config,
	}
}

// AddToBlacklist 将 Token 加入黑名单
func (b *TokenBlacklist) AddToBlacklist(ctx context.Context, jti string, ttl time.Duration) error {
	if !b.config.Enabled || b.redisClient == nil {
		return nil
	}
	
	key := b.buildKey(jti)
	return b.redisClient.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted 检查 Token 是否在黑名单中
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if !b.config.Enabled || b.redisClient == nil {
		return false, nil
	}
	
	key := b.buildKey(jti)
	_, err := b.redisClient.Get(ctx, key).Result()
	
	if err == redis.Nil {
		return false, nil
	}
	
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	
	return true, nil
}

// buildKey 构建 Redis Key
func (b *TokenBlacklist) buildKey(jti string) string {
	return fmt.Sprintf("%s%s", b.config.KeyPrefix, jti)
}

// LoginAttemptManager 登录失败管理器
type LoginAttemptManager struct {
	redisClient *redis.Client
	config      LoginSafetyConfig
}

// NewLoginAttemptManager 创建登录失败管理器
func NewLoginAttemptManager(client *redis.Client, config LoginSafetyConfig) *LoginAttemptManager {
	return &LoginAttemptManager{
		redisClient: client,
		config:      config,
	}
}

// RecordFailedAttempt 记录失败尝试
func (m *LoginAttemptManager) RecordFailedAttempt(ctx context.Context, username string) error {
	if !m.config.Enabled() || m.redisClient == nil {
		return nil
	}
	
	key := m.buildFailKey(username)
	
	// 增加失败计数
	count, err := m.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to record login attempt: %w", err)
	}
	
	// 设置过期时间
	m.redisClient.Expire(ctx, key, time.Duration(m.config.LockDuration)*time.Second)
	
	// 如果达到最大失败次数，锁定账户
	if count >= int64(m.config.MaxAttempts) {
		lockKey := m.buildLockKey(username)
		m.redisClient.Set(ctx, lockKey, "1", time.Duration(m.config.LockDuration)*time.Second)
	}
	
	return nil
}

// ClearFailedAttempts 清除失败尝试
func (m *LoginAttemptManager) ClearFailedAttempts(ctx context.Context, username string) error {
	if !m.config.Enabled() || m.redisClient == nil {
		return nil
	}
	
	key := m.buildFailKey(username)
	lockKey := m.buildLockKey(username)
	
	pipe := m.redisClient.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, lockKey)
	_, err := pipe.Exec(ctx)
	
	return err
}

// IsLocked 检查账户是否被锁定
func (m *LoginAttemptManager) IsLocked(ctx context.Context, username string) (bool, error) {
	if !m.config.Enabled() || m.redisClient == nil {
		return false, nil
	}
	
	key := m.buildLockKey(username)
	_, err := m.redisClient.Get(ctx, key).Result()
	
	if err == redis.Nil {
		return false, nil
	}
	
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}
	
	return true, nil
}

// GetFailedCount 获取失败次数
func (m *LoginAttemptManager) GetFailedCount(ctx context.Context, username string) (int, error) {
	if !m.config.Enabled() || m.redisClient == nil {
		return 0, nil
	}
	
	key := m.buildFailKey(username)
	count, err := m.redisClient.Get(ctx, key).Int()
	
	if err == redis.Nil {
		return 0, nil
	}
	
	if err != nil {
		return 0, fmt.Errorf("failed to get failed count: %w", err)
	}
	
	return count, nil
}

// buildFailKey 构建失败计数 Key
func (m *LoginAttemptManager) buildFailKey(username string) string {
	return fmt.Sprintf("auth:login:fail:%s", username)
}

// buildLockKey 构建锁定 Key
func (m *LoginAttemptManager) buildLockKey(username string) string {
	return fmt.Sprintf("auth:lock:%s", username)
}

// Enabled 检查是否启用
func (c LoginSafetyConfig) Enabled() bool {
	return c.MaxAttempts > 0 && c.LockDuration > 0
}
