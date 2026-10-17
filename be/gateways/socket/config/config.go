package socketConfig

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	GRPCTarget        string        `env:"SOCKET_GATEWAY_GRPC_TARGET" envDefault:"127.0.0.1:9093"`
	Address           string        `env:"SOCKET_GATEWAY_HTTP_ADDR" envDefault:":8083"`
	ShutdownTimeout   time.Duration `env:"SOCKET_GATEWAY_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	DependencyTimeout time.Duration `env:"SOCKET_GATEWAY_DEPENDENCY_TIMEOUT" envDefault:"5s"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
