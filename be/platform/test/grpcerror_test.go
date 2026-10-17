package test

import (
	"errors"
	"testing"

	"github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorMapperMapsMatchingRule(t *testing.T) {
	target := errors.New("invalid input")
	mapper := grpcError.NewErrorMapper(
		codes.Internal,
		"internal error",
		grpcError.MatchError(target, codes.InvalidArgument, "invalid request"),
	)

	err := mapper.Map(target)
	if status.Code(err) != codes.InvalidArgument || status.Convert(err).Message() != "invalid request" {
		t.Fatalf("mapped error = %v, want invalid argument with stable message", err)
	}
}

func TestErrorMapperUsesFallback(t *testing.T) {
	mapper := grpcError.NewErrorMapper(codes.Internal, "internal error")
	err := mapper.Map(errors.New("unexpected"))
	if status.Code(err) != codes.Internal || status.Convert(err).Message() != "internal error" {
		t.Fatalf("fallback error = %v, want internal error", err)
	}
}
