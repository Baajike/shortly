package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// slidingWindowScript is a Lua script that atomically implements a sliding-window
// rate limiter using a Redis sorted set.
//
// KEYS[1] = rate-limit key
// ARGV[1] = current Unix time in milliseconds
// ARGV[2] = window duration in milliseconds
// ARGV[3] = request limit (integer)
// ARGV[4] = unique member for this request (nanosecond timestamp string)
//
// Returns 1 if the request is allowed, 0 if it should be rejected.
var slidingWindowScript = redis.NewScript(`
local key    = KEYS[1]
local now    = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit  = tonumber(ARGV[3])
local member = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)

local count = redis.call('ZCARD', key)
if count < limit then
    redis.call('ZADD', key, now, member)
    redis.call('PEXPIRE', key, window)
    return 1
end
return 0
`)

// RateLimiter holds the Redis client and limit parameters.
type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

// NewRateLimiter creates a new sliding-window rate limiter.
func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, limit: limit, window: window}
}

// Limit returns a Gin middleware that enforces the configured rate limit per client IP.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate:shorten:%s", ip)
		now := time.Now()

		allowed, err := rl.allow(c.Request.Context(), key, now)
		if err != nil {
			// Fail open: a Redis error should not block legitimate traffic.
			c.Next()
			return
		}

		if !allowed {
			retryAfter := int(rl.window.Seconds())
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
				"request_id":  c.GetString(RequestIDKey),
			})
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allow(ctx context.Context, key string, now time.Time) (bool, error) {
	nowMs := now.UnixMilli()
	windowMs := rl.window.Milliseconds()
	member := strconv.FormatInt(now.UnixNano(), 10)

	result, err := slidingWindowScript.Run(
		ctx, rl.rdb,
		[]string{key},
		nowMs, windowMs, rl.limit, member,
	).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
