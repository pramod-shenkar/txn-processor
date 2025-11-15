package dao

import (
	"context"
	"log/slog"
	"txn-processor/config"
	"txn-processor/internal/adapter/outbound/gorm/entity"
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"
)

type Dao struct {
	port.HealthDao
	port.AccountDao
	port.TransferDao
}

var _ port.Outbound = new(Dao)

func New(ctx context.Context, db config.DB, cache config.Cache, tracer tracing.Tracer) (*Dao, error) {
	conn, err := NewConnections(db, cache, tracer)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create DB connections", "error", err)
		return nil, err
	}

	return &Dao{
		HealthDao:   NewHealthDAO(conn),
		AccountDao:  NewAccountDAO(conn),
		TransferDao: NewTransferDAO(conn),
	}, nil
}

func AutoMigrate(ctx context.Context) error {
	conn, err := GetConnections()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get DB connections", "error", err)
		return err
	}

	if err := conn.db.AutoMigrate(
		&entity.Account{},
		&entity.Transfer{},
	); err != nil {
		slog.ErrorContext(ctx, "failed to migrate entities", "error", err)
		return err
	}

	return nil
}
