package config

import (
	"context"
	"errors"
	"orderstreamrest/internal/repositories/elsearch"
	"orderstreamrest/internal/repositories/redis"
	"orderstreamrest/internal/repositories/sqlserver"
	"orderstreamrest/pkg/logger"
	"time"
)

// App - a struct that holds a redis client
type App struct {
	Redis     *redis.RedisInternal
	ES        *elsearch.Client
	Logger    *logger.Logger
	SqlServer *sqlserver.Internal
}

// NewConfig - a function that returns a new Config struct
func NewConfig() (*App, error) {

	cfg := new(App)

	err := cfg.newClientRedis()
	if err != nil {
		return cfg, err
	}

	err = cfg.newClientES()
	if err != nil {
		return cfg, err
	}

	loggerConfig := logger.Config{
		FlushInterval: 5 * time.Second,
		BufferSize:    1000,
		LogLevel:      logger.LevelDebug,
		EnableCaller:  true,
		LogDir:        "logs",
	}

	cfg.Logger = logger.NewLogger(loggerConfig)

	sqlServer, err := sqlserver.NewSQLServerInternal()
	if err != nil {
		return cfg, err
	}

	cfg.SqlServer = sqlServer

	return cfg, nil
}

// CloseAll - a function that closes all connections
func (cfg *App) CloseAll() {
	if cfg.Redis != nil {
		_ = cfg.Redis.Redis.Close()
	}

	if cfg.ES != nil {
		_ = cfg.ES.ES.Indices.Flush.WithContext(context.Background())
	}

	if cfg.Logger != nil {
		_ = cfg.Logger.Close()
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
		InsecureSkipVerify: true,
	})
	if err != nil {
		return errors.New("creating elastic client: " + err.Error())
	}

	cfg.ES = es
	return nil
}
