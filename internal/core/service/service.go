package service

import (
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"
)

type Service struct {
	port.HealthService
	port.AccountService
	port.TransferService
}

var _ port.Inbound = new(Service)

func New(dao port.Outbound, tracer tracing.Tracer) *Service {
	return &Service{
		HealthService:   NewHealthService(dao, tracer),
		AccountService:  NewAccountService(dao, tracer),
		TransferService: NewTransferService(dao, tracer),
	}
}
