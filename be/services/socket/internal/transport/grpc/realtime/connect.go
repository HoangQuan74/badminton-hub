package realtimegrpc

import (
	socketException "github.com/kainguyen/badminton-hub/be/services/socket/internal/exception"
	socketV1 "github.com/kainguyen/badminton-hub/contracts/gen/socket/v1"
)

func (s *RealtimeServer) Connect(stream socketV1.SocketService_ConnectServer) error {
	// Authenticate the first stream message before subscribing.
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	if !ValidateAccessToken(first.GetAccessToken(), s.secret) {
		return socketException.ErrInvalidAccessToken
	}

	// Forward messages between the authenticated stream and Redis.
	ctx := stream.Context()
	pubsub, err := s.subscribe(ctx)
	if err != nil {
		return err
	}
	defer pubsub.Close()
	go s.receiveMessages(ctx, stream)
	return s.sendMessages(ctx, stream, pubsub)
}
