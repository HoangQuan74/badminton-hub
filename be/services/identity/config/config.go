package identityConfig

import (
	"errors"
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Environment         string        `env:"APP_ENV" envDefault:"local"`
	GRPCAddr            string        `env:"IDENTITY_GRPC_ADDR" envDefault:":9091"`
	DatabaseURL         string        `env:"IDENTITY_DATABASE_URL,required"`
	RedisURL            string        `env:"IDENTITY_REDIS_URL" envDefault:""`
	JWTSecret           string        `env:"IDENTITY_JWT_SECRET,required"`
	AccessTokenTTL      time.Duration `env:"IDENTITY_ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenTTL     time.Duration `env:"IDENTITY_REFRESH_TOKEN_TTL" envDefault:"168h"`
	ShutdownTimeout     time.Duration `env:"IDENTITY_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	DatabasePingTimeout time.Duration `env:"IDENTITY_DATABASE_PING_TIMEOUT" envDefault:"5s"`
	DependencyTimeout   time.Duration `env:"IDENTITY_DEPENDENCY_TIMEOUT" envDefault:"5s"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("IDENTITY_JWT_SECRET must contain at least 32 characters")
	}
	return cfg, nil
}
