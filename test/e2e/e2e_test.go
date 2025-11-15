package e2e_test

import (
	"context"
	"testing"
	"time"

	"txn-processor/config"
	"txn-processor/internal/adapter/inbound/fiber/router"
	"txn-processor/internal/adapter/outbound/gorm/dao"
	"txn-processor/internal/core/service"
	"txn-processor/pkg/tracing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

type E2eSuite struct {
	suite.Suite
	app        *fiber.App
	outbound   *dao.Dao
	mariaC     *mariadb.MariaDBContainer
	redisC     *redis.RedisContainer
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (s *E2eSuite) SetupSuite() {
	s.ctx, s.cancelFunc = context.WithTimeout(context.Background(), 5*time.Minute)

	mariaC, err := mariadb.Run(s.ctx,
		"mariadb:11",
		mariadb.WithDatabase("testdb"),
		mariadb.WithUsername("user"),
		mariadb.WithPassword("password"),
	)
	s.Require().NoError(err)
	s.mariaC = mariaC

	redisC, err := redis.Run(s.ctx, "redis:7-alpine")
	s.Require().NoError(err)
	s.redisC = redisC

	mariaHost, _ := s.mariaC.Host(s.ctx)
	mariaPort, _ := s.mariaC.MappedPort(s.ctx, "3306")
	redisHost, _ := s.redisC.Host(s.ctx)
	redisPort, _ := s.redisC.MappedPort(s.ctx, "6379")

	cfg := config.DB{
		Host:     mariaHost,
		Port:     mariaPort.Port(),
		User:     "user",
		Password: "password",
		DBName:   "testdb",
		LogLevel: 1,
		Tuning: struct {
			MaxOpenConns       int `env:"DB_MAX_OPEN_CONNS" envDefault:"50"`
			MaxIdleConns       int `env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
			ConnMaxLifetimeMin int `env:"DB_CONN_MAX_LIFETIME_MIN" envDefault:"5"`
		}{
			MaxOpenConns:       10,
			MaxIdleConns:       5,
			ConnMaxLifetimeMin: 5,
		},
	}

	cacheCfg := config.Cache{
		Host: redisHost,
		Port: redisPort.Port(),
	}

	tracer := tracing.NewBlankTracer()
	outbound, err := dao.New(s.ctx, cfg, cacheCfg, tracer)
	s.Require().NoError(err)
	s.outbound = outbound

	s.Require().NoError(dao.AutoMigrate(s.ctx))

	inbound := service.New(outbound, tracer)
	s.app = router.New(inbound, tracer)
}

func (s *E2eSuite) TearDownSuite() {
	defer s.cancelFunc()

	if s.mariaC != nil {
		_ = s.mariaC.Terminate(s.ctx)
	}
	if s.redisC != nil {
		_ = s.redisC.Terminate(s.ctx)
	}
}

func (s *E2eSuite) TestLifecycle() {

}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2eSuite))
}
