package config

import (
	"time"

	"github.com/caarlos0/env/v10"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
)

type Config struct {
	S3Enabled           bool          `env:"CATALOG_S3_ENABLED" envDefault:"false"`
	S3Bucket            string        `env:"CATALOG_S3_BUCKET"`
	S3Region            string        `env:"CATALOG_S3_REGION" envDefault:"us-east-1"`
	S3Endpoint          string        `env:"CATALOG_S3_ENDPOINT"`
	S3AccessKey         string        `env:"CATALOG_S3_ACCESS_KEY"`
	S3SecretKey         string        `env:"CATALOG_S3_SECRET_KEY"`
	IdentityGRPCTarget  string        `env:"CATALOG_IDENTITY_GRPC_TARGET" envDefault:"127.0.0.1:9091"`
	GRPCAddr            string        `env:"CATALOG_GRPC_ADDR" envDefault:":9094"`
	DatabaseURL         string        `env:"CATALOG_DATABASE_URL,required"`
	ShutdownTimeout     time.Duration `env:"CATALOG_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	DatabasePingTimeout time.Duration `env:"CATALOG_DATABASE_PING_TIMEOUT" envDefault:"5s"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseURL == "" {
		return Config{}, catalogException.ErrDatabaseURLRequired
	}
	return cfg, nil
}
