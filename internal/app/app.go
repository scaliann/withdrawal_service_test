package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/scaliann/withdrawal_service_test/config"
	"github.com/scaliann/withdrawal_service_test/internal/adapter/postgres"
	"github.com/scaliann/withdrawal_service_test/internal/adapter/redis"
	"github.com/scaliann/withdrawal_service_test/internal/controller/http"
	"github.com/scaliann/withdrawal_service_test/internal/usecase"
	"github.com/scaliann/withdrawal_service_test/pkg/httpserver"
	"github.com/scaliann/withdrawal_service_test/pkg/metrics"
	pgpool "github.com/scaliann/withdrawal_service_test/pkg/postgres"
	redislib "github.com/scaliann/withdrawal_service_test/pkg/redis"
	"github.com/scaliann/withdrawal_service_test/pkg/router"
	"github.com/scaliann/withdrawal_service_test/pkg/transaction"
)

func Run(ctx context.Context, c config.Config) error {

	pgPool, err := pgpool.New(ctx, c.Postgres)
	if err != nil {
		return fmt.Errorf("postgres.New: %w", err)
	}
	defer pgPool.Close()

	transaction.Init(pgPool)

	// Metrics
	httpMetrics := metrics.NewHTTPServer()

	redisClient, err := redislib.New(c.Redis)
	if err != nil {
		return fmt.Errorf("redis.New: %w", err)
	}
	defer redisClient.Close()

	uc := usecase.New(postgres.New(), redis.NewRedis(redisClient), usecase.AuthConfig{
		AccessTTL:  time.Second * time.Duration(c.Auth.AccessTTLSeconds),
		RefreshTTL: time.Second * time.Duration(c.Auth.RefreshTTLSeconds),
		JWTSecret:  c.Auth.JWTSecret,
		JWTIssuer:  c.Auth.JWTIssuer,
	})
	r := router.New()
	http.WithdrawalRouter(r, uc, httpMetrics)
	httpServer := httpserver.New(r, c.HTTP)

	log.Info().Msg("App started!")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Info().Msg("App shutting down")

	httpServer.Close()
	log.Info().Msg("App shut down")
	return nil
}
