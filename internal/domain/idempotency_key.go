package domain

import "github.com/google/uuid"

type IdempotencyKey struct {
	BalanceID      int
	IdempotencyKey string
	PayloadHash    string
	WithdrawalID   *uuid.UUID
}
