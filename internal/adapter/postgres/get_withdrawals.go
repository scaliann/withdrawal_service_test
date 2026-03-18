package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func (p *Postgres) GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdrawal, error) {

	txOrPool := transaction.TryExtractTX(ctx)
	rows, err := txOrPool.Query(ctx, getWithdrawalsSql, userID)
	if err != nil {
		return nil, fmt.Errorf("txOrPool.Query: %w", err)
	}
	defer rows.Close()

	withdrawals := make([]domain.Withdrawal, 0)
	for rows.Next() {
		dto := struct {
			ID          uuid.UUID
			BalanceID   int
			Amount      int
			Destination string
		}{}

		err = rows.Scan(&dto.ID, &dto.BalanceID, &dto.Amount, &dto.Destination)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		withdrawals = append(withdrawals, domain.Withdrawal{
			ID:          dto.ID,
			BalanceID:   dto.BalanceID,
			Amount:      dto.Amount,
			Destination: dto.Destination,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return withdrawals, nil
}

const getWithdrawalsSql = `
SELECT w.id, w.balance_id, w.amount, w.destination 
FROM withdrawals w 
JOIN balances b ON w.balance_id = b.id
WHERE b.user_id = $1
ORDER BY w.created_at 
DESC`
