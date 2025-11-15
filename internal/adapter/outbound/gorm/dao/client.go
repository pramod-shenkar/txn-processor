package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"txn-processor/config"
	"txn-processor/pkg/tracing"

	redistracing "github.com/go-redis/redis/extra/redisotel/v8"

	gormtracing "gorm.io/plugin/opentelemetry/tracing"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Connections struct {
	db     *gorm.DB
	cache  *redis.Client
	tracer tracing.Tracer
}

var (
	connOnce sync.Once
	connInst *Connections
	connErr  error
)

// Singleton initializer
func NewConnections(dbconf config.DB, cacheconf config.Cache, tracer tracing.Tracer) (*Connections, error) {
	connOnce.Do(func() {
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbconf.User,
			dbconf.Password,
			dbconf.Host,
			dbconf.Port,
			dbconf.DBName,
		)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(dbconf.LogLevel)),
		})
		if err != nil {
			connErr = fmt.Errorf("failed to connect to mariadb: %w", err)
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			connErr = fmt.Errorf("failed to get sql.DB from gorm: %w", err)
			return
		}
		sqlDB.SetMaxOpenConns(dbconf.Tuning.MaxOpenConns)
		sqlDB.SetMaxIdleConns(dbconf.Tuning.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(dbconf.Tuning.ConnMaxLifetimeMin) * time.Minute)

		if tracer.IsEnabled() {
			if err := db.Use(gormtracing.NewPlugin()); err != nil {
				connErr = fmt.Errorf("failed to use otelgorm plugin from gorm: %w", err)
				return
			}
		}

		cache := redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", cacheconf.Host, cacheconf.Port),
			Password:     cacheconf.Password,
			DB:           cacheconf.DB,
			PoolSize:     cacheconf.Tuning.PoolSize,
			MinIdleConns: cacheconf.Tuning.MinIdleConns,
			DialTimeout:  time.Duration(cacheconf.Tuning.DialTimeoutSec) * time.Second,
			ReadTimeout:  time.Duration(cacheconf.Tuning.ReadTimeoutSec) * time.Second,
			WriteTimeout: time.Duration(cacheconf.Tuning.WriteTimeoutSec) * time.Second,
		})

		if tracer.IsEnabled() {
			cache.AddHook(redistracing.NewTracingHook())
		}

		connInst = &Connections{db: db, cache: cache, tracer: tracer}
	})

	return connInst, connErr
}

func GetConnections() (*Connections, error) {
	if connInst == nil {
		if errors.Is(connErr, nil) {
			return nil, errors.New("connections not initialized")
		}
		return nil, connErr
	}

	return connInst, nil
}

func (c *Connections) Close(ctx context.Context) error {
	var errs []error

	if c.db != nil {
		if sqlDB, err := c.db.DB(); err == nil {
			slog.InfoContext(ctx, "Closing MariaDB connection")
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("db close error: %w", err))
			}
		}
	}

	if c.cache != nil {
		slog.InfoContext(ctx, "Closing KeyDB connection")
		if err := c.cache.Close(); err != nil {
			errs = append(errs, fmt.Errorf("redis close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("connection close errors: %v", errs)
	}
	return nil
}
