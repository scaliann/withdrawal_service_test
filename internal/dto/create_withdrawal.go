package dto

import "github.com/google/uuid"

type CreateWithdrawalInput struct {
	BalanceID      int    `json:"balance_id"`
	Amount         int    `json:"amount"`
	Destination    string `json:"destination"`
	IdempotencyKey string `json:"idempotency_key"`
}

type CreateWithdrawalOutput struct {
	ID uuid.UUID `json:"id"`
}
