package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/pkg/otel/tracer"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func (p *Postgres) DebitBalance(ctx context.Context, BalanceID int, amount int) (int, error) {
	ctx, span := tracer.Start(ctx, "DebitBalance")
	defer span.End()

	txOrPool := transaction.TryExtractTX(ctx)

	var available int
	err := txOrPool.QueryRow(ctx, debitBalanceSql, amount, BalanceID).Scan(&available)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrInsufficientBalance
		}
		return 0, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return available, nil
}

const debitBalanceSql = `
UPDATE balances
SET available = available - $1
WHERE id = $2 AND available >= $1
RETURNING available;
`
