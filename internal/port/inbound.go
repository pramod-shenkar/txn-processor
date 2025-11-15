package port

import (
	"context"
	"txn-processor/internal/core/model"
)

type Inbound interface {
	HealthService
	AccountService
	TransferService
}

type HealthService interface {
	Check(ctx context.Context) error
}

type AccountService interface {
	CreateAccount(ctx context.Context, req model.AccountCreateRequest) (*model.AccountCreateResponse, error)
	GetAccount(ctx context.Context, id int64) (*model.AccountGetResponse, error)
}

type TransferService interface {
	ProcessTransfer(ctx context.Context, req model.TransferRequest) (*model.TransferResponse, error)
}
