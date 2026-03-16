package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis wraps the Redis client
type Redis struct {
	Client *redis.Client
}

// Get retrieves a value from cache
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Set sets a value in cache with TTL
func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// Delete deletes a key from cache
func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

// DeleteByPattern deletes all keys matching a pattern
func (r *Redis) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, newCursor, err := r.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.Client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.Client.Close()
}
