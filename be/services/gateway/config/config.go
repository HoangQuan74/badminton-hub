package gatewayconfig

import (
	"time"

	platformconfig "github.com/kainguyen/badminton-hub/be/platform/config"
)

type Config struct {
	Environment       string
	HTTPAddr          string
	IdentityURL       string
	ShutdownTimeout   time.Duration
	DependencyTimeout time.Duration
}

func Load() (Config, error) {
	identityURL, err := platformconfig.Required("GATEWAY_IDENTITY_URL")
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := platformconfig.Duration("GATEWAY_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	dependencyTimeout, err := platformconfig.Duration("GATEWAY_DEPENDENCY_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Environment:       platformconfig.Env("APP_ENV", "local"),
		HTTPAddr:          platformconfig.Env("GATEWAY_HTTP_ADDR", ":8080"),
		IdentityURL:       identityURL,
		ShutdownTimeout:   shutdownTimeout,
		DependencyTimeout: dependencyTimeout,
	}, nil
}
