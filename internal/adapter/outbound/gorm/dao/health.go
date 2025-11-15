package dao

import (
	"context"
	"fmt"
	"txn-processor/internal/port"
)

type HealthDAO struct {
	*Connections
}

var _ port.HealthDao = new(HealthDAO)

func NewHealthDAO(conn *Connections) *HealthDAO {
	return &HealthDAO{conn}
}

func (m *HealthDAO) Ping(ctx context.Context) error {

	ctx, span := m.tracer.Start(ctx, "dao.health.ping")
	defer span.End()

	// var result int
	// if err := m.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
	// 	return err
	// }

	db, err := m.db.WithContext(ctx).DB()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("db ping failed: %w", err)
	}

	if err := m.cache.Ping(ctx).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}
