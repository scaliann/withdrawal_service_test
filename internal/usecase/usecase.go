package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
)

type Redis interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

type Postgres interface {
	GetWithdrawal(ctx context.Context, id uuid.UUID) (domain.Withdrawal, error)
	GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdrawal, error)
	GetUserWithdrawalsSummary(
		ctx context.Context,
		from time.Time,
		to time.Time,
		limit int,
		offset int,
	) ([]domain.UserWithdrawalsSummary, error)
	CreateWithdrawal(ctx context.Context, withdrawal domain.Withdrawal) error
	DebitBalance(ctx context.Context, BalanceID int, amount int) (int, error)
	CreateIdempotencyKey(ctx context.Context, balanceID int, key string, hash string) (int, error)
	GetIdempotencyKey(ctx context.Context, balanceID int, key string) (domain.IdempotencyKey, error)
	SetIdempotencyKeyWithdrawalID(ctx context.Context, balanceID int, key string, withdrawalID uuid.UUID) error
	GetUserByCredentials(ctx context.Context, username string, password string) (domain.User, error)
}

type AuthConfig struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	JWTSecret  string
	JWTIssuer  string
}

type UseCase struct {
	postgres Postgres
	redis    Redis
	auth     AuthConfig
}

func New(postgres Postgres, redis Redis, auth AuthConfig) *UseCase {
	if auth.AccessTTL <= 0 {
		auth.AccessTTL = time.Minute * 15
	}
	if auth.RefreshTTL <= 0 {
		auth.RefreshTTL = time.Hour * 24 * 30
	}

	return &UseCase{
		postgres: postgres,
		redis:    redis,
		auth:     auth,
	}
}
