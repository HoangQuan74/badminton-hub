package grpc

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	tenantDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type Authorizer struct {
	tokens *auth.TokenManager
	staff  *staffDomain.Service
}

func NewAuthorizer(tokens *auth.TokenManager, staff *staffDomain.Service) *Authorizer {
	return &Authorizer{
		tokens: tokens,
		staff:  staff,
	}
}

// AuthorizeAdmin adapts the current admin check for endpoint authorization callbacks.
func (a *Authorizer) AuthorizeAdmin(ctx context.Context, accessToken string) error {
	_, err := a.RequireAdmin(ctx, accessToken)
	return err
}

func (a *Authorizer) RequireAdmin(ctx context.Context, accessToken string) (model.Staff, error) {
	staffID, err := a.tokens.ParseAccessToken(accessToken)
	if err != nil {
		return model.Staff{}, identityException.ErrAuthInvalidToken
	}

	value, err := a.staff.GetStaff(ctx, staffID)
	if err != nil {
		return model.Staff{}, err
	}

	if !value.IsAdmin || value.Status != constant.AccountStatusActive {
		return model.Staff{}, identityException.ErrStaffForbidden
	}

	return value, nil
}

// RequireTenantActor derives identity from the verified token and the current user record.
func RequireTenantActor(ctx context.Context, authentication *auth.Service, token string, ownerOnly bool) (string, error) {
	id, err := authentication.ParseAccessToken(token)
	if err != nil {
		return "", identityException.ErrGRPCInvalidCredentials
	}

	user, err := authentication.GetPrincipal(ctx, id)
	if err != nil {
		if err == identityException.ErrAuthUserNotFound {
			return "", identityException.ErrGRPCInvalidCredentials
		}

		return "", identityException.ErrGRPCIdentityUnavailable
	}

	if user.Status != constant.AccountStatusActive || !tenantDomain.CanAccessPortal(user) {
		return "", identityException.ErrGRPCTenantAccessDenied
	}

	if ownerOnly && user.RoleName != constant.RoleTenantOwner {
		return "", identityException.ErrGRPCTenantOwnerRequired
	}

	return user.ID, nil
}
