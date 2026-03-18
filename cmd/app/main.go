package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/scaliann/withdrawal_service_test/internal/app"
	"github.com/scaliann/withdrawal_service_test/pkg/logger"
	"github.com/scaliann/withdrawal_service_test/pkg/otel"

	"github.com/scaliann/withdrawal_service_test/config"
)

func main() {
	c, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("config.New")
	}

	logger.Init(c.Logger)

	ctx := context.Background()

	err = otel.Init(ctx, c.OTEL)
	if err != nil {
		log.Fatal().Err(err).Msg("otel.Init")
	}
	defer otel.Close()
	err = app.Run(ctx, c)

	if err != nil {
		log.Error().Err(err).Msg("app.Run")
	}

}
