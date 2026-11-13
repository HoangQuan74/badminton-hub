package identitytest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	tenantGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/tenant"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTenantCreateGRPCRejectsInvalidTokenAndDirectInvalidInput(t *testing.T) {
	repository := &tenantCreateRepository{}
	tokens := auth.NewTokenManager("test-secret-for-owner-enrollment", time.Minute, time.Hour)
	server := tenantGRPC.NewTenantServer(tenant.NewService(repository), nil, tokens, nil)
	if _, err := server.CreateOrganization(context.Background(), &identityV1.TenantCreateOrganizationRequest{AccessToken: "invalid"}); status.Code(err) != codes.Unauthenticated || repository.calls != 0 {
		t.Fatalf("authentication: %v", err)
	}
	token, _, err := tokens.NewAccessToken("00000000-0000-4000-8000-000000000003")
	if err != nil {
		t.Fatal(err)
	}
	_, err = server.CreateOrganization(context.Background(), &identityV1.TenantCreateOrganizationRequest{AccessToken: token, IdempotencyKey: "invalid"})
	if status.Code(err) != codes.InvalidArgument || repository.calls != 0 {
		t.Fatalf("direct validation: %v", err)
	}
	var reason string
	var fieldCount int
	for _, detail := range status.Convert(err).Details() {
		switch value := detail.(type) {
		case *errdetails.ErrorInfo:
			reason = value.Reason
		case *errdetails.BadRequest:
			fieldCount = len(value.FieldViolations)
		}
	}
	if reason != "INVALID_REQUEST" || fieldCount != 3 {
		t.Fatalf("reason=%s fields=%d", reason, fieldCount)
	}

	request := &identityV1.TenantCreateOrganizationRequest{AccessToken: token, Name: "Valid", Slug: "valid", IdempotencyKey: "84f6cb1a-27b1-4e9b-9a14-95e36cfbbac6"}
	for _, test := range []struct {
		err    error
		code   codes.Code
		reason string
	}{
		{identityException.ErrTenantSlugConflict, codes.AlreadyExists, "ORGANIZATION_SLUG_ALREADY_EXISTS"},
		{identityException.ErrTenantKeyConflict, codes.AlreadyExists, "IDEMPOTENCY_KEY_CONFLICT"},
		{identityException.ErrTenantRoleConflict, codes.FailedPrecondition, "TENANT_ROLE_CONFLICT"},
		{identityException.ErrTenantForbidden, codes.PermissionDenied, "FORBIDDEN"},
		{identityException.ErrTenantUnauthorized, codes.Unauthenticated, "UNAUTHORIZED"},
		{identityException.ErrTenantUnavailable, codes.Unavailable, "DEPENDENCY_UNAVAILABLE"},
		{context.DeadlineExceeded, codes.DeadlineExceeded, "DEPENDENCY_TIMEOUT"},
		{errors.New("private database error"), codes.Internal, "INTERNAL_ERROR"},
	} {
		repository.err = test.err
		_, err := server.CreateOrganization(context.Background(), request)
		if status.Code(err) != test.code {
			t.Fatalf("%v: %v", test.err, err)
		}
		info, ok := status.Convert(err).Details()[0].(*errdetails.ErrorInfo)
		if !ok || info.Reason != test.reason {
			t.Fatalf("details: %v", status.Convert(err).Details())
		}
	}
}
