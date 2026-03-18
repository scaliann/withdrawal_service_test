package postgres

import (
	"context"
	"fmt"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/pkg/otel/tracer"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func (p *Postgres) CreateWithdrawal(ctx context.Context, withdrawal domain.Withdrawal) error {
	ctx, span := tracer.Start(ctx, "adapter postgres CreateWithdrawal")
	defer span.End()

	args := []any{
		withdrawal.ID,
		withdrawal.BalanceID,
		withdrawal.Amount,
		withdrawal.Destination,
	}

	txOrPool := transaction.TryExtractTX(ctx)

	_, err := txOrPool.Exec(ctx, createWithdrawalSql, args...)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}
	return nil
}

const createWithdrawalSql = `
INSERT INTO withdrawals (id, balance_id, amount, destination) 
VALUES ($1, $2, $3, $4)
`
