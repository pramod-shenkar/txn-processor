package model

import "time"

type TransferRequest struct {
	SourceAccountID      int64  `json:"source_account_id"`
	DestinationAccountID int64  `json:"destination_account_id"`
	Amount               string `json:"amount"`
}

type TransferResponse struct {
	TransactionID        int64     `json:"transaction_id"`
	SourceAccountID      int64     `json:"source_account_id"`
	DestinationAccountID int64     `json:"destination_account_id"`
	Amount               string    `json:"amount"`
	CreatedAt            time.Time `json:"created_at"`
}
