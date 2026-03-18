package usecase

import (
	"context"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/internal/dto"
)

func (u *UseCase) GetWithdrawals(ctx context.Context, userID int) (dto.GetWithdrawalsOutput, error) {
	if userID <= 0 {
		return dto.GetWithdrawalsOutput{}, domain.ErrUnauthorized
	}

	withdrawals, err := u.postgres.GetWithdrawals(ctx, userID)
	if err != nil {
		return dto.GetWithdrawalsOutput{}, err
	}

	return dto.GetWithdrawalsOutput{
		Withdrawals: withdrawals,
	}, nil
}
