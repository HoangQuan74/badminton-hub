package exception

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrGRPCAuthenticationRequired  = status.Error(codes.Unauthenticated, "authentication required")
	ErrGRPCIdentityUnavailable     = status.Error(codes.Unavailable, "identity unavailable")
	ErrGRPCInvalidCredentials      = status.Error(codes.Unauthenticated, "invalid credentials")
	ErrGRPCTenantOwnerRequired     = status.Error(codes.PermissionDenied, "tenant owner required")
	ErrGRPCImageStorageUnavailable = status.Error(codes.Unavailable, "image storage unavailable")
)
