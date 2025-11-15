package port

import (
	"context"
	"txn-processor/internal/core/model"
)

type Outbound interface {
	HealthDao
	AccountDao
	TransferDao
}

type HealthDao interface {
	Ping(context.Context) error
}

type AccountDao interface {
	CreateAccount(ctx context.Context, req model.AccountCreateRequest) error
	GetAccountByID(ctx context.Context, id int64) (*model.AccountGetResponse, error)
}

type TransferDao interface {
	RunTransferTx(ctx context.Context, req model.TransferRequest) (*model.TransferResponse, error)
}
