package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/DostonAkhmedov/task-manager/internal/transport/response"
)

// RateLimiter implements token bucket rate limiting using Redis
type RateLimiter struct {
	client     *redis.Client
	maxTokens  int
	refillRate time.Duration
}

// NewRateLimiter creates a new rate limiter
// maxTokens: maximum requests per refillRate
// refillRate: time window for the rate limit
func NewRateLimiter(client *redis.Client, requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		client:     client,
		maxTokens:  requestsPerMinute,
		refillRate: time.Minute,
	}
}

// Middleware returns a rate limiting middleware that uses user ID or IP address as key
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user identifier (user ID from context, or IP)
			identifier := rl.getUserIdentifier(r)

			// Check rate limit
			allowed, retryAfter := rl.allow(r.Context(), identifier)
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
				response.Error(w, http.StatusTooManyRequests, 
					"rate limit exceeded - max "+strconv.Itoa(rl.maxTokens)+" requests per minute",
					"Retry after "+retryAfter.String())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// allow checks if the request should be allowed
// Returns (allowed bool, retryAfter duration)
func (rl *RateLimiter) allow(ctx context.Context, identifier string) (bool, time.Duration) {
	key := fmt.Sprintf("rate_limit:%s", identifier)

	// Increment counter
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		// If Redis error, allow request (fail open)
		return true, 0
	}

	// First request in this window - set expiration
	if count == 1 {
		_ = rl.client.Expire(ctx, key, rl.refillRate).Val()
	}

	// Check if exceeds limit
	if count > int64(rl.maxTokens) {
		ttl := rl.client.TTL(ctx, key).Val()
		retryAfter := ttl
		if retryAfter <= 0 {
			retryAfter = rl.refillRate
		}
		return false, retryAfter
	}

	return true, 0
}

// getUserIdentifier returns a unique identifier for the user/client
// Tries to get user ID from context, falls back to IP address
func (rl *RateLimiter) getUserIdentifier(r *http.Request) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, ok := r.Context().Value("user_id").(string); ok && userID != "" {
		return userID
	}

	// Fall back to IP address
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	return r.RemoteAddr
}

// Reset clears the rate limit counter for an identifier (useful for testing)
func (rl *RateLimiter) Reset(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("rate_limit:%s", identifier)
	return rl.client.Del(ctx, key).Err()
}
