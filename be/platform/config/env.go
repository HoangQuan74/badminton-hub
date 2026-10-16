package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

func Env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Required(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New(key + " is required")
	}
	return value, nil
}

func Duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return value, nil
}
