package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Lua script for rate limiting using token bucket algorithm
	luaRateLimitScript = `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])  -- max tokens
		local refill_rate = tonumber(ARGV[2])  -- tokens added per second
		local now = tonumber(ARGV[3])  -- current timestamp in seconds
		local capacity = tonumber(ARGV[4])  -- bucket capacity (same as limit in this case)

		-- Get current tokens count and last refill time
		local tokens_info = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(tokens_info[1])
		local last_refill = tonumber(tokens_info[2])

		-- Initialize if this is the first access
		if not tokens or not last_refill then
			tokens = capacity
			last_refill = now
		end

		-- Calculate tokens to add based on elapsed time
		local elapsed = now - last_refill
		local tokens_to_add = elapsed * refill_rate

		-- Update tokens count (but not exceeding capacity)
		tokens = math.min(capacity, tokens + tokens_to_add)
		last_refill = now

		-- Check if we can consume a token
		local allowed = 0
		local remaining_tokens = tokens
		if tokens >= 1 then
			tokens = tokens - 1
			allowed = 1
			remaining_tokens = tokens
		end

		-- Update Redis with new values
		redis.call('HSET', key, 'tokens', tokens, 'last_refill', now)
		redis.call('EXPIRE', key, 600)  -- Expire after 10 minutes of inactivity

		-- Calculate reset time (when the bucket will be full again)
		-- Since we don't have math.ceil in Redis Lua, we implement ceiling manually
		local remaining_capacity = capacity - tokens
		local time_to_full = remaining_capacity / refill_rate
		-- Manual ceiling: if fractional part exists, round up
		local integer_part = math.floor(time_to_full)
		local fractional_part = time_to_full - integer_part
		local reset_time = now
		if fractional_part > 0 then
			reset_time = now + integer_part + 1
		else
			reset_time = now + integer_part
		end

		-- Return: allowed flag, remaining tokens, reset timestamp
		return {allowed, remaining_tokens, reset_time}
	`
)

var rateLimitScript = redis.NewScript(luaRateLimitScript)

// DistributedRateLimiter 分布式频率限制器
type DistributedRateLimiter struct {
	client *Client
	limit  float64 // 最大请求数
	burst  int     // 突发容量
}

// NewDistributedRateLimiter 创建新的分布式频率限制器
func NewDistributedRateLimiter(client *Client, limit float64, burst int) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		client: client,
		limit:  limit,
		burst:  burst,
	}
}

// Allow 检查给定标识符是否允许请求
func (drl *DistributedRateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	now := float64(time.Now().Unix())
	refillRate := drl.limit                    // tokens added per second
	capacity := float64(drl.burst)            // bucket capacity

	result, err := rateLimitScript.Run(ctx, drl.client.Client, []string{identifier},
		capacity, refillRate, now, capacity).Result()

	if err != nil {
		return false, fmt.Errorf("rate limit script error: %v", err)
	}

	results, ok := result.([]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected result type from rate limit script")
	}

	if len(results) < 1 {
		return false, fmt.Errorf("unexpected result length from rate limit script")
	}

	allowed, ok := results[0].(int64)
	if !ok {
		return false, fmt.Errorf("unexpected allowed value type from rate limit script")
	}

	return allowed == 1, nil
}

// AllowN 检查给定标识符是否允许 n 个请求
func (drl *DistributedRateLimiter) AllowN(ctx context.Context, identifier string, n int) (bool, error) {
	// 对于简单实现，我们一次只检查一个请求
	// 更复杂的实现可以修改 Lua 脚本来处理多个请求
	for i := 0; i < n; i++ {
		allowed, err := drl.Allow(ctx, identifier)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// GetRateLimitInfo 获取频率限制信息
func (drl *DistributedRateLimiter) GetRateLimitInfo(ctx context.Context, identifier string) (allowed bool, remaining int64, resetTime int64, err error) {
	now := float64(time.Now().Unix())
	refillRate := drl.limit                    // tokens added per second
	capacity := float64(drl.burst)            // bucket capacity

	result, err := rateLimitScript.Run(ctx, drl.client.Client, []string{identifier},
		capacity, refillRate, now, capacity).Result()

	if err != nil {
		return false, 0, 0, fmt.Errorf("rate limit script error: %v", err)
	}

	results, ok := result.([]interface{})
	if !ok {
		return false, 0, 0, fmt.Errorf("unexpected result type from rate limit script")
	}

	if len(results) < 3 {
		return false, 0, 0, fmt.Errorf("unexpected result length from rate limit script")
	}

	allowedInt, err := strconv.ParseInt(fmt.Sprintf("%v", results[0]), 10, 64)
	if err != nil {
		return false, 0, 0, fmt.Errorf("error parsing allowed value: %v", err)
	}

	remainingInt, err := strconv.ParseInt(fmt.Sprintf("%v", results[1]), 10, 64)
	if err != nil {
		return false, 0, 0, fmt.Errorf("error parsing remaining value: %v", err)
	}

	resetTimeInt, err := strconv.ParseInt(fmt.Sprintf("%v", results[2]), 10, 64)
	if err != nil {
		return false, 0, 0, fmt.Errorf("error parsing reset time value: %v", err)
	}

	return allowedInt == 1, remainingInt, resetTimeInt, nil
}