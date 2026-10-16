package identityconfig

import (
	"errors"
	"time"

	platformconfig "github.com/kainguyen/badminton-hub/be/platform/config"
)

type Config struct {
	Environment         string
	HTTPAddr            string
	DatabaseURL         string
	RedisURL            string
	JWTSecret           string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	ShutdownTimeout     time.Duration
	DatabasePingTimeout time.Duration
	DependencyTimeout   time.Duration
}

func Load() (Config, error) {
	databaseURL, err := platformconfig.Required("IDENTITY_DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	jwtSecret, err := platformconfig.Required("IDENTITY_JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	if len(jwtSecret) < 32 {
		return Config{}, errors.New("IDENTITY_JWT_SECRET must contain at least 32 characters")
	}
	accessTokenTTL, err := platformconfig.Duration("IDENTITY_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTokenTTL, err := platformconfig.Duration("IDENTITY_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := platformconfig.Duration("IDENTITY_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	databasePingTimeout, err := platformconfig.Duration("IDENTITY_DATABASE_PING_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	dependencyTimeout, err := platformconfig.Duration("IDENTITY_DEPENDENCY_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Environment:         platformconfig.Env("APP_ENV", "local"),
		HTTPAddr:            platformconfig.Env("IDENTITY_HTTP_ADDR", ":8080"),
		DatabaseURL:         databaseURL,
		RedisURL:            platformconfig.Env("IDENTITY_REDIS_URL", ""),
		JWTSecret:           jwtSecret,
		AccessTokenTTL:      accessTokenTTL,
		RefreshTokenTTL:     refreshTokenTTL,
		ShutdownTimeout:     shutdownTimeout,
		DatabasePingTimeout: databasePingTimeout,
		DependencyTimeout:   dependencyTimeout,
	}, nil
}
