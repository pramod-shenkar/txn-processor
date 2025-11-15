package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"txn-processor/internal/adapter/outbound/gorm/entity"
	"txn-processor/internal/core/model"
	"txn-processor/internal/port"

	"golang.org/x/sync/singleflight"
)

const accountTTL = 60 * time.Second

type accountDAO struct {
	*Connections
	sf singleflight.Group
}

var _ port.AccountDao = (*accountDAO)(nil)

func NewAccountDAO(conn *Connections) port.AccountDao {
	return &accountDAO{
		Connections: conn,
		sf:          singleflight.Group{},
	}
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

	resp := model.AccountGetResponse{
		AccountID: e.AccountID,
		Balance:   e.Balance,
	}

	b, _ := json.Marshal(resp)
	key := fmt.Sprintf("account:%d", e.AccountID)

	if err := d.cache.Set(ctx, key, b, accountTTL).Err(); err != nil {
		span.RecordError(err)
		d.cache.Del(ctx, key).Err()
	}

	return nil
}

func (d *accountDAO) GetAccountByID(ctx context.Context, id int64) (*model.AccountGetResponse, error) {
	ctx, span := d.tracer.Start(ctx, "dao.account.get")
	defer span.End()

	key := fmt.Sprintf("account:%d", id)

	val, err := d.cache.Get(ctx, key).Result()
	if err == nil {
		var cached model.AccountGetResponse
		if json.Unmarshal([]byte(val), &cached) == nil {
			return &cached, nil
		}
		span.RecordError(fmt.Errorf("cache unmarshal error for key %s", key))
	} else {
		span.RecordError(err)
	}

	result, err, _ := d.sf.Do(key, func() (interface{}, error) {
		var e entity.Account

		if err := d.db.WithContext(ctx).
			Model(&entity.Account{}).
			Where("account_id = ?", id).
			First(&e).Error; err != nil {
			span.RecordError(err)
			return nil, err
		}

		resp := &model.AccountGetResponse{
			AccountID: e.AccountID,
			Balance:   e.Balance,
		}

		b, _ := json.Marshal(resp)
		if err := d.cache.Set(ctx, key, b, accountTTL).Err(); err != nil {
			span.RecordError(err)
			d.cache.Del(ctx, key).Err()
		}

		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.AccountGetResponse), nil
}
