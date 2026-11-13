package exception

import (
	"context"
	"errors"

	grpcError "github.com/kainguyen/badminton-hub/be/platform/grpc/error"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/protoadapt"
)

var TenantCreateGRPCMapper = grpcError.NewErrorMapper(
	codes.Internal, "owner enrollment failed",
	grpcError.MatchError(ErrAuthInvalidToken, codes.Unauthenticated, "authentication required"),
	grpcError.MatchError(ErrTenantUnauthorized, codes.Unauthenticated, "authentication required"),
	grpcError.MatchError(ErrTenantForbidden, codes.PermissionDenied, "enrollment forbidden"),
	grpcError.MatchError(ErrTenantInvalidInput, codes.InvalidArgument, "invalid enrollment input"),
	grpcError.MatchError(ErrTenantRoleConflict, codes.FailedPrecondition, "account role conflict"),
	grpcError.MatchError(ErrTenantSlugConflict, codes.AlreadyExists, "organization slug already exists"),
	grpcError.MatchError(ErrTenantKeyConflict, codes.AlreadyExists, "idempotency key conflict"),
	grpcError.MatchError(ErrTenantUnavailable, codes.Unavailable, "dependency unavailable"),
	grpcError.MatchError(context.DeadlineExceeded, codes.DeadlineExceeded, "dependency timeout"),
	grpcError.MatchError(context.Canceled, codes.Canceled, "request canceled"),
)

func TenantCreateGRPCDetails(err error) error {
	// Attach only allowlisted, non-sensitive reasons and field codes.
	reason := "INTERNAL_ERROR"
	for _, entry := range []struct {
		target error
		code   string
	}{
		{ErrTenantInvalidInput, "INVALID_REQUEST"},
		{ErrTenantUnauthorized, "UNAUTHORIZED"},
		{ErrTenantForbidden, "FORBIDDEN"},
		{ErrTenantRoleConflict, "TENANT_ROLE_CONFLICT"},
		{ErrTenantSlugConflict, "ORGANIZATION_SLUG_ALREADY_EXISTS"},
		{ErrTenantKeyConflict, "IDEMPOTENCY_KEY_CONFLICT"},
		{ErrTenantUnavailable, "DEPENDENCY_UNAVAILABLE"},
		{context.DeadlineExceeded, "DEPENDENCY_TIMEOUT"},
	} {
		if errors.Is(err, entry.target) {
			reason = entry.code
			break
		}
	}
	details := []protoadapt.MessageV1{&errdetails.ErrorInfo{Reason: reason, Domain: "badminton.identity.tenant"}}
	var validationError *TenantValidationError
	if errors.As(err, &validationError) {
		badRequest := &errdetails.BadRequest{}
		for _, field := range validationError.Fields {
			badRequest.FieldViolations = append(badRequest.FieldViolations, &errdetails.BadRequest_FieldViolation{Field: field.Field, Description: field.Code})
		}
		details = append(details, badRequest)
	}
	return grpcError.WithDetails(err, details...)
}
