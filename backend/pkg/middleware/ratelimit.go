package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig defines rate limit parameters.
type RateLimitConfig struct {
	Max    int
	Window time.Duration
}

// Lua script: atomic INCR + conditional EXPIRE. Returns [count, ttl].
var rateLimitScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("TTL", KEYS[1])
return {count, ttl}
`)

// RateLimit returns a Gin middleware that enforces per-IP rate limiting using Redis.
// Fail-closed: rejects request if Redis is unavailable.
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	windowSec := int(cfg.Window.Seconds())

	return func(c *gin.Context) {
		ctx := context.Background()
		ip := c.ClientIP()
		key := fmt.Sprintf("rl:%s:%s", c.FullPath(), ip)

		result, err := rateLimitScript.Run(ctx, rdb, []string{key}, windowSec).Int64Slice()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"status":  http.StatusServiceUnavailable,
				"message": "Service temporarily unavailable",
			})
			return
		}

		count := int(result[0])
		ttl := int(result[1])

		remaining := cfg.Max - count
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if count > cfg.Max {
			c.Header("Retry-After", fmt.Sprintf("%d", ttl+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}

// RateLimitByUser is like RateLimit but keys on authenticated userID instead of IP.
func RateLimitByUser(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	windowSec := int(cfg.Window.Seconds())

	return func(c *gin.Context) {
		ctx := context.Background()
		identity := c.GetString("userID")
		if identity == "" {
			identity = c.ClientIP()
		}
		key := fmt.Sprintf("rl:%s:%s", c.FullPath(), identity)

		result, err := rateLimitScript.Run(ctx, rdb, []string{key}, windowSec).Int64Slice()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"status":  http.StatusServiceUnavailable,
				"message": "Service temporarily unavailable",
			})
			return
		}

		count := int(result[0])
		ttl := int(result[1])

		remaining := cfg.Max - count
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if count > cfg.Max {
			c.Header("Retry-After", fmt.Sprintf("%d", ttl+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
