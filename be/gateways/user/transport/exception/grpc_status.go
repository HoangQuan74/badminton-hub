package exception

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrGRPCIdentityUnavailable = status.Error(codes.Unavailable, "identity unavailable")
	ErrGRPCTenantUserRequired  = status.Error(codes.PermissionDenied, "tenant user is required")
)
