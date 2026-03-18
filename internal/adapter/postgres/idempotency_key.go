package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/pkg/otel/tracer"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func (p *Postgres) CreateIdempotencyKey(
	ctx context.Context,
	balanceID int,
	key string,
	hash string,
) (int, error) {
	ctx, span := tracer.Start(ctx, "CreateIdempotencyKey")
	defer span.End()

	txOrPool := transaction.TryExtractTX(ctx)

	var created int
	err := txOrPool.QueryRow(ctx, createIdempotencyKeySQL, balanceID, key, hash).Scan(&created)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return created, nil
}

func (p *Postgres) GetIdempotencyKey(ctx context.Context, balanceID int, key string) (domain.IdempotencyKey, error) {
	txOrPool := transaction.TryExtractTX(ctx)

	row := domain.IdempotencyKey{}
	err := txOrPool.QueryRow(ctx, getIdempotencyKeySQL, balanceID, key).Scan(
		&row.BalanceID,
		&row.IdempotencyKey,
		&row.PayloadHash,
		&row.WithdrawalID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.IdempotencyKey{}, domain.ErrNotFound
		}
		return domain.IdempotencyKey{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return row, nil
}

func (p *Postgres) SetIdempotencyKeyWithdrawalID(
	ctx context.Context,
	balanceID int,
	key string,
	withdrawalID uuid.UUID,
) error {
	txOrPool := transaction.TryExtractTX(ctx)

	cmd, err := txOrPool.Exec(ctx, setIdempotencyKeyWithdrawalIDSQL, withdrawalID, balanceID, key)
	if err != nil {
		return fmt.Errorf("txOrPool.Exec: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

const createIdempotencyKeySQL = `
INSERT INTO idempotency_keys (balance_id, idempotency_key, payload_hash)
VALUES ($1, $2, $3)
ON CONFLICT (balance_id, idempotency_key) DO NOTHING
RETURNING balance_id;
`

const getIdempotencyKeySQL = `
SELECT balance_id, idempotency_key, payload_hash, withdrawal_id
FROM idempotency_keys
WHERE balance_id = $1 AND idempotency_key = $2;
`

const setIdempotencyKeyWithdrawalIDSQL = `
UPDATE idempotency_keys
SET withdrawal_id = $1
WHERE balance_id = $2 AND idempotency_key = $3;
`
