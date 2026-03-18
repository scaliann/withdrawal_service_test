package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/scaliann/withdrawal_service_test/pkg/httpserver"
	"github.com/scaliann/withdrawal_service_test/pkg/logger"
	"github.com/scaliann/withdrawal_service_test/pkg/otel"
	"github.com/scaliann/withdrawal_service_test/pkg/postgres"
	"github.com/scaliann/withdrawal_service_test/pkg/redis"
)

type App struct {
	Name    string `envconfig:"APP_NAME" required:"true"`
	Version string `envconfig:"APP_VERSION" required:"true"`
}

type Auth struct {
	JWTSecret         string `envconfig:"AUTH_JWT_SECRET" required:"true"`
	JWTIssuer         string `envconfig:"AUTH_JWT_ISSUER" default:"withdrawal-service"`
	AccessTTLSeconds  int    `envconfig:"AUTH_ACCESS_TTL_SEC" default:"900"`
	RefreshTTLSeconds int    `envconfig:"AUTH_REFRESH_TTL_SEC" default:"2592000"`
}

type Config struct {
	App      App
	Auth     Auth
	Logger   logger.Config
	OTEL     otel.Config
	Postgres postgres.Config
	Redis    redis.Config
	HTTP     httpserver.Config
}

func New() (Config, error) {
	var config Config

	err := godotenv.Load(".env")
	if err != nil {
		return config, fmt.Errorf("godotenv.Load: %w", err)
	}

	err = envconfig.Process("", &config)
	if err != nil {
		return config, fmt.Errorf("envconfig.Process: %w", err)
	}

	return config, nil
}
