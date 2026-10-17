package realtimegrpc

import (
	"context"

	socketV1 "github.com/kainguyen/badminton-hub/contracts/gen/socket/v1"
	"github.com/redis/go-redis/v9"
)

func (s *Server) subscribe(ctx context.Context) (*redis.PubSub, error) {
	pubsub := s.redis.Subscribe(ctx, s.channel)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}
	return pubsub, nil
}

func (s *Server) receiveMessages(ctx context.Context, stream socketV1.SocketService_ConnectServer) {
	for {
		message, err := stream.Recv()
		if err != nil {
			return
		}
		if len(message.GetPayload()) == 0 {
			continue
		}
		_ = s.redis.Publish(ctx, s.channel, message.GetPayload()).Err()
	}
}

func (s *Server) sendMessages(ctx context.Context, stream socketV1.SocketService_ConnectServer, pubsub *redis.PubSub) error {
	for {
		message, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			return err
		}
		if err := stream.Send(&socketV1.ServerMessage{Payload: []byte(message.Payload)}); err != nil {
			return err
		}
	}
}
