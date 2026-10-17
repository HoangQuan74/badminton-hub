package config

import (
	"errors"
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	GRPCAddr          string        `env:"SOCKET_GRPC_ADDR" envDefault:":9093"`
	JWTSecret         string        `env:"SOCKET_JWT_SECRET,required"`
	RedisURL          string        `env:"SOCKET_REDIS_URL" envDefault:"redis://127.0.0.1:16379/2"`
	RedisChannel      string        `env:"SOCKET_REDIS_CHANNEL" envDefault:"badminton:realtime"`
	DependencyTimeout time.Duration `env:"SOCKET_DEPENDENCY_TIMEOUT" envDefault:"5s"`
	ShutdownTimeout   time.Duration `env:"SOCKET_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("SOCKET_JWT_SECRET must contain at least 32 characters")
	}
	return cfg, nil
}
