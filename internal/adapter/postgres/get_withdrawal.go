package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func (p *Postgres) GetWithdrawal(ctx context.Context, id uuid.UUID) (domain.Withdrawal, error) {

	dto := struct {
		ID          uuid.UUID
		BalanceID   int
		Amount      int
		Destination string
	}{}

	dest := []any{
		&dto.ID,
		&dto.BalanceID,
		&dto.Amount,
		&dto.Destination,
	}

	txOrPool := transaction.TryExtractTX(ctx)

	err := txOrPool.QueryRow(ctx, getWithdrawalSql, id).Scan(dest...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Withdrawal{}, domain.ErrNotFound
		}
		return domain.Withdrawal{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	withdrawal := domain.Withdrawal{
		ID:          dto.ID,
		BalanceID:   dto.BalanceID,
		Amount:      dto.Amount,
		Destination: dto.Destination,
	}

	return withdrawal, nil
}

const getWithdrawalSql = `
SELECT id, balance_id, amount, destination
FROM withdrawals
WHERE id = $1
`
