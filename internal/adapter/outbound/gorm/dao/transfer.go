package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"txn-processor/internal/adapter/outbound/gorm/entity"
	"txn-processor/internal/core/model"
	"txn-processor/internal/port"
	"txn-processor/pkg/decimal"

	"gorm.io/gorm/clause"
)

const (
	transferTTL = 60 * time.Second
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

	if err := tx.Model(&entity.Account{}).
		Clauses(LockClause).
		Where("account_id = ?", req.SourceAccountID).
		First(&source).Error; err != nil {
		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Model(&entity.Account{}).
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

	if err := tx.Model(&entity.Account{}).
		Where("id = ?", source.ID).
		Updates(map[string]interface{}{"balance": source.Balance}).Error; err != nil {
		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Model(&entity.Account{}).
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

	if err := tx.Model(&entity.Transfer{}).
		Create(&record).Error; err != nil {
		span.RecordError(err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		span.RecordError(err)
		return nil, err
	}

	resp := &model.TransferResponse{
		TransactionID:        int64(record.ID),
		SourceAccountID:      record.SourceAccountID,
		DestinationAccountID: record.DestinationAccountID,
		Amount:               record.Amount,
		CreatedAt:            record.CreatedAt,
	}

	trKey := fmt.Sprintf("transfer:%d", record.ID)
	b, _ := json.Marshal(resp)
	if err := d.cache.Set(ctx, trKey, b, transferTTL).Err(); err != nil {
		span.RecordError(err)
		_ = d.cache.Del(ctx, trKey).Err()
	}

	srcKey := fmt.Sprintf("account:%d", source.AccountID)
	src := model.AccountGetResponse{AccountID: source.AccountID, Balance: source.Balance}
	b1, _ := json.Marshal(src)
	if err := d.cache.Set(ctx, srcKey, b1, accountTTL).Err(); err != nil {
		span.RecordError(err)
		_ = d.cache.Del(ctx, srcKey).Err()
	}

	dstKey := fmt.Sprintf("account:%d", dest.AccountID)
	dst := model.AccountGetResponse{AccountID: dest.AccountID, Balance: dest.Balance}
	b2, _ := json.Marshal(dst)
	if err := d.cache.Set(ctx, dstKey, b2, accountTTL).Err(); err != nil {
		span.RecordError(err)
		_ = d.cache.Del(ctx, dstKey).Err()
	}

	return resp, nil
}
