package dao

import (
	"context"
	"errors"
	"txn-processor/internal/adapter/outbound/gorm/entity"
	"txn-processor/internal/core/model"
	"txn-processor/internal/port"
	"txn-processor/pkg/decimal"

	"gorm.io/gorm/clause"
)

var LockClause = clause.Locking{Strength: "UPDATE"}

type transferDAO struct {
	*Connections
}

var _ port.TransferDao = (*transferDAO)(nil)

func NewTransferDAO(conn *Connections) port.TransferDao {
	return &transferDAO{Connections: conn}
}

func (d *transferDAO) RunTransferTx(ctx context.Context, req model.TransferRequest) (*model.TransferResponse, error) {
	ctx, span := d.tracer.Start(ctx, "dao.transfer.tx")
	defer span.End()

	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		span.RecordError(tx.Error)
		return nil, tx.Error
	}

	var source entity.Account
	var dest entity.Account

	if err := tx.
		Model(&entity.Account{}).
		Clauses(LockClause).
		Where("account_id = ?", req.SourceAccountID).
		First(&source).Error; err != nil {

		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.
		Model(&entity.Account{}).
		Clauses(LockClause).
		Where("account_id = ?", req.DestinationAccountID).
		First(&dest).Error; err != nil {

		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	if decimal.LessThan(source.Balance, req.Amount) {
		tx.Rollback()
		err := errors.New("insufficient balance")
		span.RecordError(err)
		return nil, err
	}

	source.Balance = decimal.Sub(source.Balance, req.Amount)
	dest.Balance = decimal.Add(dest.Balance, req.Amount)

	if err := tx.
		Model(&entity.Account{}).
		Where("id = ?", source.ID).
		Updates(map[string]interface{}{"balance": source.Balance}).Error; err != nil {

		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.
		Model(&entity.Account{}).
		Where("id = ?", dest.ID).
		Updates(map[string]interface{}{"balance": dest.Balance}).Error; err != nil {

		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	record := entity.Transfer{
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               req.Amount,
	}

	if err := tx.
		Model(&entity.Transfer{}).
		Create(&record).Error; err != nil {

		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &model.TransferResponse{
		TransactionID:        int64(record.ID),
		SourceAccountID:      record.SourceAccountID,
		DestinationAccountID: record.DestinationAccountID,
		Amount:               record.Amount,
		CreatedAt:            record.CreatedAt,
	}, nil
}
