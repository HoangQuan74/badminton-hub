package test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
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

func TestErrorMapperPreservesOptionalDetailsAndFallback(t *testing.T) {
	target := errors.New("business conflict")
	detailed := grpcError.WithDetails(fmt.Errorf("wrapped: %w", target), &errdetails.ErrorInfo{Reason: "CONFLICT", Domain: "test"})
	if !errors.Is(detailed, target) {
		t.Fatal("details must preserve the cause")
	}
	mapper := grpcError.NewErrorMapper(codes.Internal, "safe error", grpcError.MatchError(target, codes.AlreadyExists, "conflict"))
	mapped := mapper.Map(detailed)
	info, ok := status.Convert(mapped).Details()[0].(*errdetails.ErrorInfo)
	if status.Code(mapped) != codes.AlreadyExists || !ok || info.Reason != "CONFLICT" {
		t.Fatalf("details lost: %v", mapped)
	}
	if mapper.Map(nil) != nil || grpcError.WithDetails(nil) != nil {
		t.Fatal("nil must remain nil")
	}
	fallback := grpcError.NewErrorMapper(codes.Internal, "safe fallback").Map(detailed)
	if status.Code(fallback) != codes.Internal || len(status.Convert(fallback).Details()) != 1 {
		t.Fatalf("fallback: %v", fallback)
	}
	// gRPC rejects details attached to OK; retain the original nil status safely.
	if grpcError.NewErrorMapper(codes.OK, "ok").Map(detailed) != nil {
		t.Fatal("OK fallback changed")
	}
}

func TestErrorMapperUsesFallback(t *testing.T) {
	mapper := grpcError.NewErrorMapper(codes.Internal, "internal error")
	err := mapper.Map(errors.New("unexpected"))
	if status.Code(err) != codes.Internal || status.Convert(err).Message() != "internal error" {
		t.Fatalf("fallback error = %v, want internal error", err)
	}
}
