package config

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
)

type App struct {
	Name     string `env:"APP_NAME" envDefault:"txn-processor"`
	Port     string `env:"APP_PORT" envDefault:"9999"`
	LogLevel int    `env:"APP_LOG_LEVEL" envDefault:"-4"`
	Env      string `env:"APP_ENV" envDefault:"default"`
	DB       DB
	Cache    Cache
	Otel     Otel
}

type DB struct {
	Host          string `env:"DB_HOST" envDefault:"localhost"`
	Port          string `env:"DB_PORT" envDefault:"3306"`
	User          string `env:"DB_USER" envDefault:"user"`
	Password      string `env:"DB_PASSWORD" envDefault:"password"`
	DBName        string `env:"DB_Name" envDefault:"default"`
	LogLevel      int    `env:"DB_LOG_LEVEL" envDefault:"4"`
	IsAutoMigrate bool   `env:"DB_IS_AUTO_MIGRATE" envDefault:"true"`

	Tuning struct {
		MaxOpenConns       int `env:"DB_MAX_OPEN_CONNS" envDefault:"50"`
		MaxIdleConns       int `env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
		ConnMaxLifetimeMin int `env:"DB_CONN_MAX_LIFETIME_MIN" envDefault:"5"`
	} `envPrefix:"DB_"`
}

type Cache struct {
	Host     string `env:"CACHE_HOST" envDefault:"localhost"`
	Port     string `env:"CACHE_PORT" envDefault:"6379"`
	User     string `env:"CACHE_USER" envDefault:""`
	Password string `env:"CACHE_PASSWORD" envDefault:""`
	DB       int    `env:"CACHE_DB" envDefault:"0"`

	Tuning struct {
		PoolSize        int `env:"CACHE_POOL_SIZE" envDefault:"20"`
		MinIdleConns    int `env:"CACHE_MIN_IDLE_CONNS" envDefault:"5"`
		DialTimeoutSec  int `env:"CACHE_DIAL_TIMEOUT_SEC" envDefault:"5"`
		ReadTimeoutSec  int `env:"CACHE_READ_TIMEOUT_SEC" envDefault:"3"`
		WriteTimeoutSec int `env:"CACHE_WRITE_TIMEOUT_SEC" envDefault:"3"`
	} `envPrefix:"CACHE_"`
}

type Otel struct {
	Metrics Metrics
	Tracer  Tracer
	Logger  Logger
}

type Metrics struct {
	IsEnabled bool `env:"OTEL_METRICS_ENABLED" envDefault:"false"`
}
type Tracer struct {
	IsEnabled bool   `env:"OTEL_TRACER_ENABLED" envDefault:"true"`
	Endpoint  string `env:"OTEL_TRACER_ENDPOINT" envDefault:"http://localhost:14268/api/traces"`
}
type Logger struct {
	IsEnabled bool `env:"OTEL_LOGGER_ENABLED" envDefault:"false"`
}

func New() *App {
	cfg := &App{}
	if err := env.Parse(cfg); err != nil {
		slog.Error("error while parsing config", "err", err)
	}
	return cfg
}
