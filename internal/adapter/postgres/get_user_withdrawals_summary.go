package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func (p *Postgres) GetUserWithdrawalsSummary(
	ctx context.Context,
	from time.Time,
	to time.Time,
	limit int,
	offset int,
) ([]domain.UserWithdrawalsSummary, error) {
	txOrPool := transaction.TryExtractTX(ctx)
	rows, err := txOrPool.Query(ctx, getUserWithdrawalsSummarySQL, from, to, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("txOrPool.Query: %w", err)
	}
	defer rows.Close()

	result := make([]domain.UserWithdrawalsSummary, 0)
	for rows.Next() {
		dto := struct {
			UserID           int
			Username         string
			WithdrawalsCount int64
			TotalAmount      string
			LastWithdrawalAt sql.NullTime
		}{}

		err = rows.Scan(
			&dto.UserID,
			&dto.Username,
			&dto.WithdrawalsCount,
			&dto.TotalAmount,
			&dto.LastWithdrawalAt,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		var lastWithdrawalAt *time.Time
		if dto.LastWithdrawalAt.Valid {
			t := dto.LastWithdrawalAt.Time
			lastWithdrawalAt = &t
		}

		result = append(result, domain.UserWithdrawalsSummary{
			UserID:           dto.UserID,
			Username:         dto.Username,
			WithdrawalsCount: dto.WithdrawalsCount,
			TotalAmount:      dto.TotalAmount,
			LastWithdrawalAt: lastWithdrawalAt,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return result, nil
}

const getUserWithdrawalsSummarySQL = `
SELECT
    u.id AS user_id,
    u.username,
    COALESCE(x.withdrawals_count, 0) AS withdrawals_count,
    COALESCE(x.total_amount, 0)::text AS total_amount,
    x.last_withdrawal_at
FROM users u
LEFT JOIN (
    SELECT
        b.user_id,
        COUNT(w.id) AS withdrawals_count,
        SUM(w.amount) AS total_amount,
        MAX(w.created_at) AS last_withdrawal_at
    FROM balances b
    JOIN withdrawals w ON w.balance_id = b.id
    WHERE w.created_at >= $1
      AND w.created_at < $2
    GROUP BY b.user_id
) x ON x.user_id = u.id
ORDER BY COALESCE(x.total_amount, 0) DESC
LIMIT $3 OFFSET $4;
`
