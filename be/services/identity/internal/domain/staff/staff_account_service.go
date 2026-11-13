package staff

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
)

type AccountService struct {
	repository StaffAccountRepository
	tokens     *auth.TokenManager
}

func NewAccountService(repository StaffAccountRepository, tokens *auth.TokenManager) *AccountService {
	return &AccountService{
		repository: repository,
		tokens:     tokens,
	}
}

func (s *AccountService) Login(ctx context.Context, email, password string) (model.Staff, model.TokenPair, error) {
	credentials, err := s.repository.FindStaffCredentialsByEmail(ctx, strings.TrimSpace(email))
	if err != nil || !security.ComparePassword(credentials.PasswordHash, password) {
		return model.Staff{}, model.TokenPair{}, identityException.ErrAuthInvalidCredentials
	}

	if credentials.Staff.Status != constant.AccountStatusActive {
		return model.Staff{}, model.TokenPair{}, identityException.ErrAuthUserInactive
	}

	// Prepare tokens only after the staff account has been admitted.
	tokens, session, err := s.tokens.NewSession(credentials.Staff.ID)
	if err != nil {
		return model.Staff{}, model.TokenPair{}, err
	}

	// Persist the hashed session before exposing credentials to the caller.
	if err := s.repository.CreateStaffSession(ctx, session); err != nil {
		return model.Staff{}, model.TokenPair{}, err
	}

	return credentials.Staff, tokens, nil
}

func (s *AccountService) Refresh(ctx context.Context, token string) (model.TokenPair, error) {
	if token == "" {
		return model.TokenPair{}, identityException.ErrAuthInvalidToken
	}

	raw, hash, expiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return model.TokenPair{}, err
	}

	userID, err := s.repository.RotateStaffSession(
		ctx,
		auth.HashRefreshToken(token),
		model.Session{
			ID:        uuid.NewString(),
			TokenHash: hash,
			ExpiresAt: expiry,
		},
	)
	if err != nil {
		return model.TokenPair{}, err
	}

	access, accessExpiry, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{
		AccessToken:      access,
		RefreshToken:     raw,
		AccessExpiresAt:  accessExpiry,
		RefreshExpiresAt: expiry,
	}, nil
}

func (s *AccountService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return identityException.ErrAuthInvalidToken
	}

	return s.repository.RevokeStaffSession(ctx, auth.HashRefreshToken(token))
}

func (s *AccountService) ChangePassword(ctx context.Context, userID, current, next string) error {
	hash, err := s.repository.GetStaffPasswordHash(ctx, userID)
	if err != nil {
		return err
	}

	if !security.ComparePassword(hash, current) {
		return identityException.ErrAuthInvalidCredentials
	}

	nextHash, err := security.HashPassword(next)
	if err != nil {
		return err
	}

	return s.repository.UpdateStaffPassword(
		ctx,
		userID,
		nextHash,
	)
}
