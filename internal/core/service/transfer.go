package service

import (
	"context"
	"strings"
	"txn-processor/internal/core/model"
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"
)

type transferService struct {
	dao    port.TransferDao
	tracer tracing.Tracer
}

var _ port.TransferService = (*transferService)(nil)

func NewTransferService(dao port.TransferDao, tracer tracing.Tracer) port.TransferService {
	return &transferService{dao: dao, tracer: tracer}
}

func (s *transferService) ProcessTransfer(ctx context.Context, req model.TransferRequest) (*model.TransferResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.transfer.process")
	defer span.End()

	if req.SourceAccountID <= 0 ||
		req.DestinationAccountID <= 0 ||
		strings.TrimSpace(req.Amount) == "" {
		err := ErrValidation
		span.RecordError(err)
		return nil, err
	}

	if req.SourceAccountID == req.DestinationAccountID {
		err := ErrValidation
		span.RecordError(err)
		return nil, err
	}

	result, err := s.dao.RunTransferTx(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &model.TransferResponse{
		TransactionID:        result.TransactionID,
		SourceAccountID:      result.SourceAccountID,
		DestinationAccountID: result.DestinationAccountID,
		Amount:               result.Amount,
		CreatedAt:            result.CreatedAt,
	}, nil
}
