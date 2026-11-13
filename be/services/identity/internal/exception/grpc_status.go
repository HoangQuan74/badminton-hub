package exception

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrGRPCInvalidCredentials  = status.Error(codes.Unauthenticated, "invalid credentials")
	ErrGRPCIdentityUnavailable = status.Error(codes.Unavailable, "identity unavailable")
	ErrGRPCTenantAccessDenied  = status.Error(codes.PermissionDenied, "tenant access denied")
	ErrGRPCTenantOwnerRequired = status.Error(codes.PermissionDenied, "tenant owner required")
)
