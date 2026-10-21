package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	IdentityGRPCTarget string        `env:"ADMIN_GATEWAY_IDENTITY_GRPC_TARGET" envDefault:"127.0.0.1:9091"`
	Address            string        `env:"ADMIN_GATEWAY_HTTP_ADDR" envDefault:":8082"`
	CORSAllowedOrigins string        `env:"ADMIN_GATEWAY_CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:5173,http://localhost:5174,http://localhost:5175,http://127.0.0.1:3000,http://127.0.0.1:5173,http://127.0.0.1:5174,http://127.0.0.1:5175"`
	ShutdownTimeout    time.Duration `env:"ADMIN_GATEWAY_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	DependencyTimeout  time.Duration `env:"ADMIN_GATEWAY_DEPENDENCY_TIMEOUT" envDefault:"5s"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
