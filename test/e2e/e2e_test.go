package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"txn-processor/config"
	"txn-processor/internal/adapter/inbound/fiber/router"
	"txn-processor/internal/adapter/outbound/gorm/dao"
	"txn-processor/internal/core/model"
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
	// Step 1: Create Account 1001 with 500
	acc1 := model.AccountCreateRequest{
		AccountID:      1001,
		InitialBalance: "500",
	}

	body1, err := json.Marshal(acc1)
	s.Require().NoError(err)

	req := httptest.NewRequest("POST", "/v1/accounts", bytes.NewReader(body1))
	req.Header.Set("Content-Type", "application/json")

	res, err := s.app.Test(req, -1)
	s.Require().NoError(err)
	s.Require().Equal(201, res.StatusCode)

	// Step 2: Create Account 2002 with 200
	acc2 := model.AccountCreateRequest{
		AccountID:      2002,
		InitialBalance: "200",
	}

	body2, err := json.Marshal(acc2)
	s.Require().NoError(err)

	req = httptest.NewRequest("POST", "/v1/accounts", bytes.NewReader(body2))
	req.Header.Set("Content-Type", "application/json")

	res, err = s.app.Test(req, -1)
	s.Require().NoError(err)
	s.Require().Equal(201, res.StatusCode)

	// Step 3: Fetch Account 1001 before transfer
	req = httptest.NewRequest("GET", "/v1/accounts/1001", nil)
	res, err = s.app.Test(req, -1)
	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode)

	var beforeAcc model.AccountGetResponse
	err = json.NewDecoder(res.Body).Decode(&beforeAcc)
	s.Require().NoError(err)
	s.Require().Equal(int64(1001), beforeAcc.AccountID)
	s.Require().Equal("500", beforeAcc.Balance)

	// Step 4: Perform Transfer 1001 -> 2002 amount=150
	transfer := model.TransferRequest{
		SourceAccountID:      1001,
		DestinationAccountID: 2002,
		Amount:               "150",
	}

	body3, err := json.Marshal(transfer)
	s.Require().NoError(err)

	req = httptest.NewRequest("POST", "/v1/transfers", bytes.NewReader(body3))
	req.Header.Set("Content-Type", "application/json")

	res, err = s.app.Test(req, -1)
	s.Require().NoError(err)
	s.Require().Equal(201, res.StatusCode)

	var tr model.TransferResponse
	err = json.NewDecoder(res.Body).Decode(&tr)
	s.Require().NoError(err)
	s.Require().Equal(int64(1001), tr.SourceAccountID)
	s.Require().Equal(int64(2002), tr.DestinationAccountID)
	s.Require().Equal("150", tr.Amount)

	// Step 5: Fetch updated Account 1001
	req = httptest.NewRequest("GET", "/v1/accounts/1001", nil)
	res, err = s.app.Test(req, -1)
	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode)

	var accAfter1 model.AccountGetResponse
	err = json.NewDecoder(res.Body).Decode(&accAfter1)
	s.Require().NoError(err)
	s.Require().Equal("350", accAfter1.Balance)

	// Step 6: Fetch updated Account 2002
	req = httptest.NewRequest("GET", "/v1/accounts/2002", nil)
	res, err = s.app.Test(req, -1)
	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode)

	var accAfter2 model.AccountGetResponse
	err = json.NewDecoder(res.Body).Decode(&accAfter2)
	s.Require().NoError(err)
	s.Require().Equal("350", accAfter2.Balance)
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2eSuite))
}
