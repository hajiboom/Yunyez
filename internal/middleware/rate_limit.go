/*
Package middleware provides middleware functions for the Gin framework,
used to handle authentication, logging, recovery, and other pre-request processing tasks.

The rate limiting middleware supports both local and distributed modes:
- Local mode: Uses in-memory storage, suitable for single-instance deployments
- Distributed mode: Uses Redis for coordination, suitable for multi-instance deployments
*/
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"yunyez/internal/common/tools"
	"yunyez/internal/pkg/logger"
	"yunyez/internal/pkg/redis"
)

// LocalRateLimiter implements rate limiting using in-memory storage.
// This is suitable for single-instance deployments.
type LocalRateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	limit    rate.Limit // Request rate limit (requests per second)
	burst    int        // Burst capacity (max tokens in bucket)
}

// Visitor represents a visitor with their associated rate limiter and last seen time.
type Visitor struct {
	limiter  *rate.Limiter // Token bucket rate limiter for this visitor
	lastSeen time.Time     // Last time this visitor made a request
}

// NewLocalRateLimiter creates a new local rate limiter with the given configuration.
// It starts a background goroutine to clean up inactive visitors periodically.
func NewLocalRateLimiter(limit rate.Limit, burst int) *LocalRateLimiter {
	rl := &LocalRateLimiter{
		visitors: make(map[string]*Visitor),
		limit:    limit,
		burst:    burst,
	}

	// Start a goroutine to clean up visitors that haven't been active for 3 minutes
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors removes visitors that haven't made requests in the last 3 minutes.
// This prevents memory leaks from accumulating inactive IP addresses.
func (rl *LocalRateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// GetLimiter retrieves or creates a rate limiter for the given IP address.
// It updates the last seen time for the visitor.
func (rl *LocalRateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.limit, rl.burst)
		rl.visitors[ip] = &Visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// DistributedRateLimiterWrapper wraps the Redis-based distributed rate limiter
// to provide a consistent interface with the local rate limiter.
type DistributedRateLimiterWrapper struct {
	drl *redis.DistributedRateLimiter // Underlying Redis-based rate limiter
}

// NewDistributedRateLimiterWrapper creates a new wrapper for the distributed rate limiter.
func NewDistributedRateLimiterWrapper(drl *redis.DistributedRateLimiter) *DistributedRateLimiterWrapper {
	return &DistributedRateLimiterWrapper{
		drl: drl,
	}
}

// Allow checks if a request from the given IP is allowed under the rate limits.
// It uses Redis to coordinate rate limiting across multiple instances.
// If Redis is unavailable, it logs an error and allows the request to pass
// to prevent a Redis failure from bringing down the entire system.
func (drw *DistributedRateLimiterWrapper) Allow(ctx context.Context, ip string) bool {
	// Use a prefixed key to store rate limit data in Redis
	allowed, err := drw.drl.Allow(ctx, fmt.Sprintf("rate_limit:%s", ip))
	if err != nil {
		// Log the error but allow the request to pass to avoid system-wide outage
		// if Redis is unavailable
		logger.Error(ctx, "Failed to check rate limit in Redis", map[string]interface{}{
			"ip":  ip,
			"err": err.Error(),
		})
		return true
	}
	return allowed
}

// RateLimitMode defines the operational mode for rate limiting.
type RateLimitMode string

const (
	// LocalMode stores rate limit data in memory, suitable for single-instance deployments.
	LocalMode RateLimitMode = "local"

	// DistributedMode stores rate limit data in Redis, suitable for multi-instance deployments.
	DistributedMode RateLimitMode = "distributed"
)

// RateLimitConfig holds the configuration for rate limiting.
type RateLimitConfig struct {
	Mode  RateLimitMode // Operational mode (local or distributed)
	Limit rate.Limit    // Request rate limit (requests per second)
	Burst int           // Burst capacity (max tokens in bucket)
}

// RateLimitMiddleware implements rate limiting middleware that supports both local and distributed modes.
// In distributed mode, it uses Redis to coordinate rate limiting across multiple service instances.
// The middleware adds trace ID to the request context for logging correlation.
//
// Parameters:
//   - config: Configuration specifying the rate limiting mode, limit, and burst capacity
//
// Returns:
//   - A Gin middleware function that enforces rate limits
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	switch config.Mode {
	case DistributedMode:
		// Initialize Redis client for distributed rate limiting
		redisClient, err := redis.NewClient()
		if err != nil {
			// Panic if Redis initialization fails, as distributed mode requires Redis
			panic(fmt.Sprintf("Failed to initialize Redis client for rate limiting: %v", err))
		}

		// Create distributed rate limiter using Redis
		distributedLimiter := redis.NewDistributedRateLimiter(redisClient, float64(config.Limit), config.Burst)
		distributedWrapper := NewDistributedRateLimiterWrapper(distributedLimiter)

		return func(c *gin.Context) {
			ip := c.ClientIP()

			// Create a context with trace ID for logging correlation
			ctx := c.Request.Context()
			traceID := tools.GetTraceID(ctx)
			ctxWithTrace := tools.WithTraceID(ctx, traceID)

			// Check if the request is allowed under rate limits
			if !distributedWrapper.Allow(ctxWithTrace, ip) {
				// Log rate limit violation with relevant context
				logger.Warn(ctxWithTrace, "Rate limit exceeded", map[string]interface{}{
					"ip":    ip,
					"limit": config.Limit,
					"burst": config.Burst,
				})

				// Return 429 Too Many Requests response
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too Many Requests",
				})
				c.Abort()
				return
			}

			// Proceed to next middleware/handler if request is allowed
			c.Next()
		}
	default: // LocalMode
		// Use local rate limiter for single-instance deployments
		localLimiter := NewLocalRateLimiter(config.Limit, config.Burst)

		return func(c *gin.Context) {
			ip := c.ClientIP()

			// Create a context with trace ID for logging correlation
			ctx := c.Request.Context()
			traceID := tools.GetTraceID(ctx)
			ctxWithTrace := tools.WithTraceID(ctx, traceID)

			// Get or create rate limiter for this IP
			limiter := localLimiter.GetLimiter(ip)

			// Check if the request is allowed under rate limits
			if !limiter.Allow() {
				// Log rate limit violation with relevant context
				logger.Warn(ctxWithTrace, "Rate limit exceeded", map[string]interface{}{
					"ip":    ip,
					"limit": config.Limit,
					"burst": config.Burst,
				})

				// Return 429 Too Many Requests response
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too Many Requests",
				})
				c.Abort()
				return
			}

			// Proceed to next middleware/handler if request is allowed
			c.Next()
		}
	}
}