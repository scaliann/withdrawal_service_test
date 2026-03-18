package domain

import "time"

type UserWithdrawalsSummary struct {
	UserID           int
	Username         string
	WithdrawalsCount int64
	TotalAmount      string
	LastWithdrawalAt *time.Time
}
