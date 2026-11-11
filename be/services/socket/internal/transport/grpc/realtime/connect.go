package realtimegrpc

import (
	"errors"

	socketV1 "github.com/kainguyen/badminton-hub/contracts/gen/socket/v1"
)

func (s *Server) Connect(stream socketV1.SocketService_ConnectServer) error {
	// Authenticate the first stream message before subscribing.
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	if !ValidateAccessToken(first.GetAccessToken(), s.secret) {
		return errors.New("invalid access token")
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
