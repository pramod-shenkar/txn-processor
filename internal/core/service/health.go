package service

import (
	"context"
	"log/slog"
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"
)

type Health struct {
	HealthDAO port.HealthDao
	tracer    tracing.Tracer
}

var _ port.HealthService = new(Health)

func NewHealthService(healthDAO port.HealthDao, tracer tracing.Tracer) port.HealthService {
	return &Health{
		HealthDAO: healthDAO,
		tracer:    tracer,
	}
}

func (s *Health) Check(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "service.health.check")
	defer span.End()

	if err := s.HealthDAO.Ping(ctx); err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Health check failed", "error", err)
		return err
	}
	return nil
}
