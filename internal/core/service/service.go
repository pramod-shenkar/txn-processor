package service

import (
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"
)

type Service struct {
}

var _ port.Inbound = new(Service)

func New(dao port.Outbound, tracer tracing.Tracer) *Service {
	return &Service{}
}
