package auth

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

// TenantService enforces tenant access while reusing credential and session handling.
type TenantService struct {
	credentials *Service
}

func NewTenantService(repository Repository, tokens *TokenManager) *TenantService {
	return &TenantService{credentials: NewService(repository, tokens)}
}

func (s *TenantService) Login(ctx context.Context, input LoginInput) (model.User, TokenPair, error) {
	// Enforce the tenant boundary before issuing a session.
	input.Portal = "tenant"
	return s.credentials.Login(ctx, input)
}

func (s *TenantService) Refresh(ctx context.Context, token string) (TokenPair, error) {
	return s.credentials.Refresh(ctx, token, "tenant")
}

func (s *TenantService) Logout(ctx context.Context, token string) error {
	return s.credentials.Logout(ctx, token)
}

func (s *TenantService) Me(ctx context.Context, accessToken string) (model.User, error) {
	// Verify the token before loading the tenant operator.
	userID, err := s.credentials.ParseAccessToken(accessToken)
	if err != nil {
		return model.User{}, err
	}
	user, err := s.credentials.Me(ctx, userID)
	if err != nil {
		return model.User{}, err
	}
	if !IsAllowedPortal(user, "tenant") {
		return model.User{}, ErrWrongPortal
	}
	return user, nil
}
