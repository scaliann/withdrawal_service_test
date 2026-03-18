package dto

import (
	"time"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
)

type GetUserWithdrawalsSummaryInput struct {
	From   string
	To     string
	Limit  string
	Offset string
}

type UserWithdrawalsSummary struct {
	UserID           int        `json:"user_id"`
	Username         string     `json:"username"`
	WithdrawalsCount int64      `json:"withdrawals_count"`
	TotalAmount      string     `json:"total_amount"`
	LastWithdrawalAt *time.Time `json:"last_withdrawal_at,omitempty"`
}

type GetUserWithdrawalsSummaryOutput struct {
	Items []UserWithdrawalsSummary `json:"items"`
}

func NewUserWithdrawalsSummaryOutput(items []domain.UserWithdrawalsSummary) GetUserWithdrawalsSummaryOutput {
	result := make([]UserWithdrawalsSummary, 0, len(items))
	for _, item := range items {
		result = append(result, UserWithdrawalsSummary{
			UserID:           item.UserID,
			Username:         item.Username,
			WithdrawalsCount: item.WithdrawalsCount,
			TotalAmount:      item.TotalAmount,
			LastWithdrawalAt: item.LastWithdrawalAt,
		})
	}

	return GetUserWithdrawalsSummaryOutput{Items: result}
}
