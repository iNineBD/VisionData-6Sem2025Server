package redis

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

// RedisInternal is a struct that contains a Redis client and a mutex
type RedisInternal struct {
	Redis *redis.Client
}

var mu sync.Mutex

// NewRedisInternal is a function that returns a new RedisInternal struct
func NewRedisInternal() (*RedisInternal, error) {

	mu = sync.Mutex{}

	// Create a new Redis client

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("connecting to Redis: %w", err)
	}

	return &RedisInternal{
		Redis: rdb,
	}, nil
}
