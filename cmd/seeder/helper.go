package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPasswordHash = "$2a$06$ZH6L4rx5op4wQ7lo3N9dfemcb1IRd8AVM4Ut.vhJWl3PYmFuZFUky" // password123
)

type currency struct {
	Code string
	Name string
}

func popularCurrencies() []currency {
	return []currency{
		{Code: "USDT", Name: "Tether"},
		{Code: "BTC", Name: "Bitcoin"},
		{Code: "ETH", Name: "Ethereum"},
		{Code: "USDC", Name: "USD Coin"},
		{Code: "BNB", Name: "BNB"},
		{Code: "XRP", Name: "XRP"},
		{Code: "ADA", Name: "Cardano"},
		{Code: "SOL", Name: "Solana"},
		{Code: "DOGE", Name: "Dogecoin"},
		{Code: "TRX", Name: "TRON"},
		{Code: "TON", Name: "Toncoin"},
		{Code: "AVAX", Name: "Avalanche"},
		{Code: "DOT", Name: "Polkadot"},
		{Code: "MATIC", Name: "Polygon"},
		{Code: "LTC", Name: "Litecoin"},
		{Code: "BCH", Name: "Bitcoin Cash"},
		{Code: "XLM", Name: "Stellar"},
		{Code: "LINK", Name: "Chainlink"},
		{Code: "ETC", Name: "Ethereum Classic"},
		{Code: "ATOM", Name: "Cosmos"},
	}
}

func runSeed(ctx context.Context, pool *pgxpool.Pool, cfg SeederConfig) error {
	if cfg.Reset {
		if err := resetData(ctx, pool); err != nil {
			return fmt.Errorf("resetData: %w", err)
		}
	}

	currencyIDs, err := ensureCurrencies(ctx, pool)
	if err != nil {
		return fmt.Errorf("ensureCurrencies: %w", err)
	}
	if len(currencyIDs) < cfg.BalancesPerUser {
		return fmt.Errorf("not enough currencies: need %d, got %d", cfg.BalancesPerUser, len(currencyIDs))
	}

	nextUserID, nextBalanceID, err := resolveStartIDs(ctx, pool)
	if err != nil {
		return fmt.Errorf("resolveStartIDs: %w", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	totalBatches := (cfg.Users + cfg.BatchSize - 1) / cfg.BatchSize
	usersLeft := cfg.Users

	for batch := 1; usersLeft > 0; batch++ {
		batchUsers := cfg.BatchSize
		if usersLeft < batchUsers {
			batchUsers = usersLeft
		}

		userRows := make([][]any, 0, batchUsers)
		balanceRows := make([][]any, 0, batchUsers*cfg.BalancesPerUser)
		withdrawalRows := make([][]any, 0, batchUsers*cfg.WithdrawalsPerUser)

		for i := 0; i < batchUsers; i++ {
			userID := nextUserID
			nextUserID++

			username := fmt.Sprintf("user_%d", userID)
			email := fmt.Sprintf("user_%d@example.com", userID)
			role := "user"
			if userID%100000 == 0 {
				role = "admin"
			}

			createdAt := randomPastTime(rng, 365*24*time.Hour)
			userRows = append(userRows, []any{
				userID,
				username,
				email,
				defaultPasswordHash,
				role,
				true,
				createdAt,
				createdAt,
			})

			userBalanceIDs := make([]int64, 0, cfg.BalancesPerUser)
			selectedCurrencies := pickDistinctCurrencies(rng, currencyIDs, cfg.BalancesPerUser)
			for _, currencyID := range selectedCurrencies {
				balanceID := nextBalanceID
				nextBalanceID++
				userBalanceIDs = append(userBalanceIDs, balanceID)

				available := randomAmount(rng, 50, 20000)
				now := time.Now().UTC()
				balanceRows = append(balanceRows, []any{
					balanceID,
					userID,
					currencyID,
					available,
					now,
					now,
				})
			}

			for w := 0; w < cfg.WithdrawalsPerUser; w++ {
				balanceID := userBalanceIDs[rng.Intn(len(userBalanceIDs))]
				amount := randomAmount(rng, 1, 250)
				withdrawalRows = append(withdrawalRows, []any{
					uuid.New(),
					balanceID,
					amount,
					fmt.Sprintf("wallet_user_%d_%d", userID, w+1),
					randomPastTime(rng, 180*24*time.Hour),
				})
			}
		}

		if err = copyUsers(ctx, pool, userRows); err != nil {
			return fmt.Errorf("copyUsers batch %d/%d: %w", batch, totalBatches, err)
		}
		if err = copyBalances(ctx, pool, balanceRows); err != nil {
			return fmt.Errorf("copyBalances batch %d/%d: %w", batch, totalBatches, err)
		}
		if err = copyWithdrawals(ctx, pool, withdrawalRows); err != nil {
			return fmt.Errorf("copyWithdrawals batch %d/%d: %w", batch, totalBatches, err)
		}

		usersLeft -= batchUsers
		if batch%cfg.LogEveryBatches == 0 || usersLeft == 0 {
			fmt.Printf(
				"[seeder] batch %d/%d done, users inserted: %d/%d\n",
				batch, totalBatches, cfg.Users-usersLeft, cfg.Users,
			)
		}
	}

	return nil
}

func resetData(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
TRUNCATE TABLE idempotency_keys, withdrawals, balances, users RESTART IDENTITY CASCADE;
`
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func ensureCurrencies(ctx context.Context, pool *pgxpool.Pool) ([]int16, error) {
	for _, c := range popularCurrencies() {
		const upsert = `
INSERT INTO currencies (code, name)
VALUES ($1, $2)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name;
`
		_, err := pool.Exec(ctx, upsert, c.Code, c.Name)
		if err != nil {
			return nil, fmt.Errorf("upsert currency %s: %w", c.Code, err)
		}
	}

	rows, err := pool.Query(ctx, `SELECT id FROM currencies ORDER BY id ASC LIMIT 20`)
	if err != nil {
		return nil, fmt.Errorf("pool.Query: %w", err)
	}
	defer rows.Close()

	ids := make([]int16, 0, 20)
	for rows.Next() {
		var id int16
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return ids, nil
}

func resolveStartIDs(ctx context.Context, pool *pgxpool.Pool) (int64, int64, error) {
	var maxUserID int64
	err := pool.QueryRow(ctx, `SELECT COALESCE(MAX(id), 0) FROM users`).Scan(&maxUserID)
	if err != nil {
		return 0, 0, fmt.Errorf("max user id: %w", err)
	}

	var maxBalanceID int64
	err = pool.QueryRow(ctx, `SELECT COALESCE(MAX(balance_id), 0) FROM balances`).Scan(&maxBalanceID)
	if err != nil {
		return 0, 0, fmt.Errorf("max balance id: %w", err)
	}

	return maxUserID + 1, maxBalanceID + 1, nil
}

func copyUsers(ctx context.Context, pool *pgxpool.Pool, rows [][]any) error {
	_, err := pool.CopyFrom(
		ctx,
		pgx.Identifier{"users"},
		[]string{"id", "username", "email", "password_hash", "role", "is_active", "created_at", "updated_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("pool.CopyFrom users: %w", err)
	}

	return nil
}

func copyBalances(ctx context.Context, pool *pgxpool.Pool, rows [][]any) error {
	_, err := pool.CopyFrom(
		ctx,
		pgx.Identifier{"balances"},
		[]string{"balance_id", "user_id", "currency_id", "available", "created_at", "updated_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("pool.CopyFrom balances: %w", err)
	}

	return nil
}

func copyWithdrawals(ctx context.Context, pool *pgxpool.Pool, rows [][]any) error {
	_, err := pool.CopyFrom(
		ctx,
		pgx.Identifier{"withdrawals"},
		[]string{"id", "balance_id", "amount", "destination", "created_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("pool.CopyFrom withdrawals: %w", err)
	}

	return nil
}

func pickDistinctCurrencies(rng *rand.Rand, ids []int16, count int) []int16 {
	perm := rng.Perm(len(ids))
	result := make([]int16, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, ids[perm[i]])
	}

	return result
}

func randomPastTime(rng *rand.Rand, maxAge time.Duration) time.Time {
	seconds := rng.Int63n(int64(maxAge.Seconds()) + 1)
	return time.Now().UTC().Add(-time.Duration(seconds) * time.Second)
}

func randomAmount(rng *rand.Rand, min int64, max int64) string {
	if max <= min {
		return fmt.Sprintf("%d.00000000", min)
	}
	value := rng.Int63n(max-min+1) + min
	return fmt.Sprintf("%d.00000000", value)
}
