package redis

import (
	"context"
	"fmt"
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
		Addr: "redis:6379",
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {

		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		if _, err := rdb.Ping(context.Background()).Result(); err != nil {
			return nil, fmt.Errorf("connecting to Redis: %w", err)
		}
	}

	return &RedisInternal{
		Redis: rdb,
	}, nil
}
