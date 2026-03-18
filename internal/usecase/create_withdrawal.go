package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/internal/dto"
	"github.com/scaliann/withdrawal_service_test/internal/idempotency"
	"github.com/scaliann/withdrawal_service_test/pkg/otel/tracer"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func (u *UseCase) CreateWithdrawal(ctx context.Context, input dto.CreateWithdrawalInput) (dto.CreateWithdrawalOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase CreateProfile")
	defer span.End()

	var output dto.CreateWithdrawalOutput
	if input.BalanceID <= 0 {
		return output, domain.ErrInvalidBalanceID
	}
	if input.Amount <= 0 {
		return output, domain.ErrInvalidAmount
	}
	if strings.TrimSpace(input.Destination) == "" {
		return output, domain.ErrDestinationRequired
	}
	if strings.TrimSpace(input.IdempotencyKey) == "" {
		return output, domain.ErrIdempotencyKeyRequired
	}

	input.Destination = strings.TrimSpace(input.Destination)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)

	withdrawal := domain.NewWithdrawal(input.BalanceID, input.Amount, input.Destination)

	payloadHash := idempotency.BuildPayloadHash(input)
	err := transaction.Wrap(ctx, func(ctx context.Context) error {
		created, err := u.postgres.CreateIdempotencyKey(ctx, input.BalanceID, input.IdempotencyKey, payloadHash)
		if err != nil {
			return fmt.Errorf("postgres.CreateIdempotencyKey: %w", err)
		}

		if created == 0 {
			idemKey, err := u.postgres.GetIdempotencyKey(ctx, input.BalanceID, input.IdempotencyKey)
			if err != nil {
				return fmt.Errorf("postgres.GetIdempotencyKey: %w", err)
			}
			if idemKey.PayloadHash != payloadHash {
				return domain.ErrIdempotencyKeyPayloadDiff
			}
			if idemKey.WithdrawalID == nil {
				return domain.ErrIdempotencyKeyAlreadyExist
			}

			output.ID = *idemKey.WithdrawalID
			return nil
		}

		err = u.postgres.CreateWithdrawal(ctx, withdrawal)
		if err != nil {
			return fmt.Errorf("postgres.CreateWithdrawal: %w", err)
		}

		_, err = u.postgres.DebitBalance(ctx, input.BalanceID, input.Amount)
		if err != nil {
			return fmt.Errorf("postgres.DebitBalance: %w", err)
		}

		err = u.postgres.SetIdempotencyKeyWithdrawalID(ctx, input.BalanceID, input.IdempotencyKey, withdrawal.ID)
		if err != nil {
			return fmt.Errorf("postgres.SetIdempotencyKeyWithdrawalID: %w", err)
		}

		output.ID = withdrawal.ID
		return nil
	})

	if err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	return output, err

}
