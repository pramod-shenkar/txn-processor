package dao

import (
	"context"
	"txn-processor/internal/adapter/outbound/gorm/entity"
	"txn-processor/internal/core/model"
	"txn-processor/internal/port"
)

type accountDAO struct {
	*Connections
}

var _ port.AccountDao = (*accountDAO)(nil)

func NewAccountDAO(conn *Connections) port.AccountDao {
	return &accountDAO{Connections: conn}
}

func (d *accountDAO) CreateAccount(ctx context.Context, req model.AccountCreateRequest) error {
	ctx, span := d.tracer.Start(ctx, "dao.account.create")
	defer span.End()

	e := entity.Account{
		AccountID: req.AccountID,
		Balance:   req.InitialBalance,
	}

	if err := d.db.WithContext(ctx).
		Model(&entity.Account{}).
		Create(&e).Error; err != nil {

		span.RecordError(err)
		return err
	}

	return nil
}

func (d *accountDAO) GetAccountByID(ctx context.Context, id int64) (*model.AccountGetResponse, error) {
	ctx, span := d.tracer.Start(ctx, "dao.account.get")
	defer span.End()

	var e entity.Account

	if err := d.db.WithContext(ctx).
		Model(&entity.Account{}).
		Where("account_id = ?", id).
		First(&e).Error; err != nil {

		span.RecordError(err)
		return nil, err
	}

	return &model.AccountGetResponse{
		AccountID: e.AccountID,
		Balance:   e.Balance,
	}, nil
}
