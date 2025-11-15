// account_service.go
package service

import (
	"context"
	"errors"
	"strings"
	"txn-processor/internal/core/model"
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation failed")
	ErrConflict   = errors.New("conflict")
)

type accountService struct {
	dao    port.AccountDao
	tracer tracing.Tracer
}

var _ port.AccountService = (*accountService)(nil)

func NewAccountService(dao port.AccountDao, tracer tracing.Tracer) port.AccountService {
	return &accountService{dao: dao, tracer: tracer}
}

func (s *accountService) CreateAccount(ctx context.Context, req model.AccountCreateRequest) (*model.AccountCreateResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.account.create")
	defer span.End()

	if req.AccountID <= 0 || strings.TrimSpace(req.InitialBalance) == "" {
		err := ErrValidation
		span.RecordError(err)
		return nil, err
	}

	if err := s.dao.CreateAccount(ctx, req); err != nil {
		span.RecordError(err)
		if isUnique(err) {
			return nil, ErrConflict
		}
		return nil, err
	}

	return &model.AccountCreateResponse{AccountID: req.AccountID}, nil
}

func (s *accountService) GetAccount(ctx context.Context, id int64) (*model.AccountGetResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.account.get")
	defer span.End()

	if id <= 0 {
		err := ErrValidation
		span.RecordError(err)
		return nil, err
	}

	acc, err := s.dao.GetAccountByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, ErrNotFound
	}

	return &model.AccountGetResponse{
		AccountID: acc.AccountID,
		Balance:   acc.Balance,
	}, nil
}

func isUnique(err error) bool {
	if err == nil {
		return false
	}
	l := strings.ToLower(err.Error())
	return strings.Contains(l, "duplicate") || strings.Contains(l, "unique")
}
