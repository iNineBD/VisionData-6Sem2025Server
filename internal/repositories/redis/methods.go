package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get is a function that returns the value of a key
func (r *RedisInternal) Get(ctx context.Context, key string) *redis.StringCmd {
	mu.Lock()
	defer mu.Unlock()
	return r.Redis.Get(ctx, key)
}

// Set is a function that sets a key value pair
func (r *RedisInternal) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	mu.Lock()
	defer mu.Unlock()
	return r.Redis.Set(ctx, key, value, expiration)
}

// Expire is a function that sets a key expiration time
func (r *RedisInternal) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	mu.Lock()
	defer mu.Unlock()
	return r.Redis.Expire(ctx, key, expiration)
}

// FlushAll is a function that flushes all keys
func (r *RedisInternal) FlushAll(ctx context.Context) *redis.StatusCmd {
	mu.Lock()
	defer mu.Unlock()
	return r.Redis.FlushAll(ctx)
}

// TTL is a function that returns the time to live of a key
func (r *RedisInternal) TTL(ctx context.Context, key string) *redis.DurationCmd {
	mu.Lock()
	defer mu.Unlock()
	return r.Redis.TTL(ctx, key)
}

// Incr is a function that increments a key
func (r *RedisInternal) Incr(ctx context.Context, key string) *redis.IntCmd {
	mu.Lock()
	defer mu.Unlock()
	return r.Redis.Incr(ctx, key)
}
