package dto

import "github.com/scaliann/withdrawal_service_test/internal/domain"

type GetWithdrawalInput struct {
	ID string `json:"id"`
}

type GetWithdrawalOutput struct {
	Withdrawal domain.Withdrawal
}

type GetWithdrawalsOutput struct {
	Withdrawals []domain.Withdrawal `json:"withdrawals"`
}
