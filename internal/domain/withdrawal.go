package domain

import (
	"github.com/google/uuid"
)

type WithdrawalStatus string

type Withdrawal struct {
	ID          uuid.UUID `json:"id"`
	BalanceID   int       `json:"balance_id"`
	Amount      int       `json:"amount"`
	Destination string    `json:"destination"`
}

func NewWithdrawal(balanceID int, amount int, destination string) Withdrawal {
	return Withdrawal{
		ID:          uuid.New(),
		BalanceID:   balanceID,
		Amount:      amount,
		Destination: destination,
	}
}
