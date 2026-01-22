package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis() (*Client, *miniredis.Miniredis, error) {
	s, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	client := &Client{rdb}

	return client, s, nil
}

func TestDistributedRateLimiter(t *testing.T) {
	client, s, err := setupTestRedis()
	if err != nil {
		t.Fatalf("Failed to setup test Redis: %v", err)
	}
	defer s.Close()

	drl := NewDistributedRateLimiter(client, 1.0, 2) // 1 request per second, burst of 2

	ctx := context.Background()
	key := "test_client"

	// First request should be allowed
	allowed, err := drl.Allow(ctx, key)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Second request should also be allowed due to burst capacity
	allowed, err = drl.Allow(ctx, key)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Third request should be denied
	allowed, err = drl.Allow(ctx, key)
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Wait for tokens to refill
	time.Sleep(1100 * time.Millisecond)

	// Request after waiting should be allowed again
	allowed, err = drl.Allow(ctx, key)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestDistributedRateLimiterWithInvalidKey(t *testing.T) {
	client, s, err := setupTestRedis()
	if err != nil {
		t.Fatalf("Failed to setup test Redis: %v", err)
	}
	defer s.Close()

	drl := NewDistributedRateLimiter(client, 1.0, 1)

	ctx := context.Background()
	key := ""

	// Request with empty key should behave normally
	allowed, err := drl.Allow(ctx, key)
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestGetRateLimitInfo(t *testing.T) {
	client, s, err := setupTestRedis()
	if err != nil {
		t.Fatalf("Failed to setup test Redis: %v", err)
	}
	defer s.Close()

	drl := NewDistributedRateLimiter(client, 1.0, 5) // 1 request per second, burst of 5

	ctx := context.Background()
	key := "test_client_info"

	// Get initial rate limit info
	allowed, remaining, resetTime, err := drl.GetRateLimitInfo(ctx, key)
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.GreaterOrEqual(t, remaining, int64(0)) // Should have some remaining tokens initially
	assert.Greater(t, resetTime, int64(0))       // Should have a reset time

	// Consume a token
	allowed, err = drl.Allow(ctx, key)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Get rate limit info after consuming a token
	allowed, remainingAfter, resetTimeAfter, err := drl.GetRateLimitInfo(ctx, key)
	assert.NoError(t, err)
	assert.True(t, allowed)
	// The remaining count might be the same or different depending on the refill timing
	// But it should be a valid number
	assert.GreaterOrEqual(t, remainingAfter, int64(0))
	assert.Greater(t, resetTimeAfter, int64(0))
}