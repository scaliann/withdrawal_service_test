package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

const getUserByCredentialsSQL = `
SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
FROM users
WHERE username = $1
  AND is_active = true
  AND password_hash = crypt($2, password_hash)
LIMIT 1;
`

func (p *Postgres) GetUserByCredentials(ctx context.Context, username string, password string) (domain.User, error) {
	txOrPool := transaction.TryExtractTX(ctx)

	var user domain.User
	err := txOrPool.QueryRow(ctx, getUserByCredentialsSQL, username, password).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}

		return domain.User{}, fmt.Errorf("txOrPool.QueryRow: %w", err)
	}

	return user, nil
}
