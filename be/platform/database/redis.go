package database

import (
	"context"
	"fmt"
	"time"

	redisClient "github.com/redis/go-redis/v9"
)

func OpenRedis(ctx context.Context, rawURL string, timeout time.Duration) (*redisClient.Client, error) {
	options, err := redisClient.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}
	client := redisClient.NewClient(options)
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}
	return client, nil
}
