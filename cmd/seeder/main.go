package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SeederConfig struct {
	DBURL              string
	Users              int
	BalancesPerUser    int
	WithdrawalsPerUser int
	BatchSize          int
	Reset              bool
	LogEveryBatches    int
}

func main() {
	cfg := loadConfig()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		panic(fmt.Errorf("pgxpool.New: %w", err))
	}
	defer pool.Close()

	started := time.Now()
	fmt.Printf(
		"[seeder] start users=%d balances/user=%d withdrawals/user=%d batch=%d reset=%v\n",
		cfg.Users, cfg.BalancesPerUser, cfg.WithdrawalsPerUser, cfg.BatchSize, cfg.Reset,
	)

	err = runSeed(ctx, pool, cfg)
	if err != nil {
		panic(fmt.Errorf("runSeed: %w", err))
	}

	fmt.Printf("[seeder] done in %s\n", time.Since(started))
}

func loadConfig() SeederConfig {
	cfg := SeederConfig{
		DBURL:              getEnv("SEED_DB_URL", "postgres://default:1234@localhost:5446/withdrawal?sslmode=disable"),
		Users:              getEnvAsInt("SEED_USERS", 1_000_000),
		BalancesPerUser:    getEnvAsInt("SEED_BALANCES_PER_USER", 2),
		WithdrawalsPerUser: getEnvAsInt("SEED_WITHDRAWALS_PER_USER", 5),
		BatchSize:          getEnvAsInt("SEED_BATCH_SIZE", 5000),
		Reset:              getEnvAsBool("SEED_RESET", true),
		LogEveryBatches:    getEnvAsInt("SEED_LOG_EVERY_BATCHES", 5),
	}

	if cfg.Users <= 0 {
		panic("SEED_USERS must be > 0")
	}
	if cfg.BalancesPerUser <= 0 {
		panic("SEED_BALANCES_PER_USER must be > 0")
	}
	if cfg.WithdrawalsPerUser <= 0 {
		panic("SEED_WITHDRAWALS_PER_USER must be > 0")
	}
	if cfg.BatchSize <= 0 {
		panic("SEED_BATCH_SIZE must be > 0")
	}
	if cfg.LogEveryBatches <= 0 {
		cfg.LogEveryBatches = 1
	}

	return cfg
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		panic(fmt.Sprintf("%s must be int: %v", key, err))
	}

	return value
}

func getEnvAsBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		panic(fmt.Sprintf("%s must be bool: %v", key, err))
	}

	return value
}
