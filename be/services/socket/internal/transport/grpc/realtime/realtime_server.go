package realtimegrpc

import (
	socketV1 "github.com/kainguyen/badminton-hub/contracts/gen/socket/v1"
	"github.com/redis/go-redis/v9"
)

type RealtimeServer struct {
	socketV1.UnimplementedSocketServiceServer
	redis   *redis.Client
	channel string
	secret  []byte
}

func NewRealtimeServer(client *redis.Client, channel, secret string) *RealtimeServer {
	return &RealtimeServer{
		redis:   client,
		channel: channel,
		secret:  []byte(secret),
	}
}
