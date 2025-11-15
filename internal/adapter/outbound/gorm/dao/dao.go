package dao

import (
	"context"
	"log/slog"
	"txn-processor/config"
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"
)

type Dao struct {
}

var _ port.Outbound = new(Dao)

func New(ctx context.Context, db config.DB, cache config.Cache, tracer tracing.Tracer) (*Dao, error) {
	_, err := NewConnections(db, cache, tracer)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create DB connections", "error", err)
		return nil, err
	}

	return &Dao{}, nil
}

func AutoMigrate(ctx context.Context) error {
	conn, err := GetConnections()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get DB connections", "error", err)
		return err
	}

	if err := conn.db.AutoMigrate(); err != nil {
		slog.ErrorContext(ctx, "failed to migrate entities", "error", err)
		return err
	}

	return nil
}
