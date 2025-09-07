package config

import (
	"context"
	"errors"
	"orderstreamrest/internal/repositories/elsearch"
	"orderstreamrest/internal/repositories/redis"
	"orderstreamrest/pkg/logger"
	"time"

	"github.com/google/uuid"
)

// App - a struct that holds a redis client
type App struct {
	Redis  *redis.RedisInternal
	ES     *elsearch.Client
	Logger *logger.ElasticsearchLogger
	// Mongo *mongo.MongoInternal
}

// NewConfig - a function that returns a new Config struct
func NewConfig() (*App, error) {

	cfg := new(App)

	executionID := uuid.New().String()[0:5]

	err := cfg.newClientRedis()
	if err != nil {
		return cfg, err
	}

	err = cfg.newClientES()
	if err != nil {
		return cfg, err
	}

	loggerConfig := logger.Config{
		Service:         "datavision-api",
		Version:         "1.0.0",
		Environment:     "homol", // or "development", "staging"
		IndexName:       "datavision-api-logs",
		FlushInterval:   1 * time.Second,
		BatchSize:       5,
		BufferSize:      1000,
		LogLevel:        logger.LevelInfo,
		EnableCaller:    true,
		EnableBody:      true, // Set to true if you want to log request/response bodies
		MaxBodySize:     1024,
		SensitiveFields: []string{"password", "token", "secret"},
		ExecutionID:     executionID,
	}

	cfg.Logger = logger.NewLogger(cfg.ES.ES, loggerConfig)

	return cfg, nil
}

// CloseAll - a function that closes all connections
func (cfg *App) CloseAll() {
	if cfg.Redis != nil {
		cfg.Redis.Redis.Close()
	}

	if cfg.ES != nil {
		cfg.ES.ES.Indices.Flush.WithContext(context.Background())
	}

	if cfg.Logger != nil {
		cfg.Logger.Close()
	}

}

// newClientRedis is a function that returns a new Redis client
func (cfg *App) newClientRedis() error {

	r, err := redis.NewRedisInternal()
	if err != nil {
		return errors.New("creating redis client: " + err.Error())
	}

	cfg.Redis = r

	return nil
}

func (cfg *App) newClientES() error {
	es, err := elsearch.NewClient(&elsearch.Config{
		MaxRetries:         3,
		RetryBackoff:       3,
		Timeout:            5 * time.Second,
		EnableLogging:      true,
		InsecureSkipVerify: false,
	})
	if err != nil {
		return errors.New("creating elastic client: " + err.Error())
	}

	cfg.ES = es
	return nil
}
