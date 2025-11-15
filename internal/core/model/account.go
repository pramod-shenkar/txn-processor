package model

type AccountCreateRequest struct {
	AccountID      int64  `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

type AccountCreateResponse struct {
	AccountID int64 `json:"account_id"`
}

type AccountGetResponse struct {
	AccountID int64  `json:"account_id"`
	Balance   string `json:"balance"`
}
