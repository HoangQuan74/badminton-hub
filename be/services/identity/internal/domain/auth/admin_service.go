package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
)

type AdminService struct {
	repository AdminRepository
	tokens     *TokenManager
}

func NewAdminService(repository AdminRepository, tokens *TokenManager) *AdminService {
	return &AdminService{repository: repository, tokens: tokens}
}

func (s *AdminService) Login(ctx context.Context, email, password string) (model.Staff, TokenPair, error) {
	credentials, err := s.repository.FindAdminByEmail(ctx, strings.TrimSpace(email))
	if err != nil || !security.ComparePassword(credentials.PasswordHash, password) {
		return model.Staff{}, TokenPair{}, ErrInvalidCredentials
	}
	if credentials.Staff.Status != "active" {
		return model.Staff{}, TokenPair{}, ErrUserInactive
	}
	access, accessExpiry, err := s.tokens.NewAccessToken(credentials.Staff.ID)
	if err != nil {
		return model.Staff{}, TokenPair{}, err
	}
	raw, hash, expiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return model.Staff{}, TokenPair{}, err
	}
	err = s.repository.CreateSession(ctx, Session{ID: uuid.NewString(), UserID: credentials.Staff.ID, TokenHash: hash, ExpiresAt: expiry})
	if err != nil {
		return model.Staff{}, TokenPair{}, err
	}
	return credentials.Staff, TokenPair{AccessToken: access, RefreshToken: raw, AccessExpiresAt: accessExpiry, RefreshExpiresAt: expiry}, nil
}

func (s *AdminService) Refresh(ctx context.Context, token string) (TokenPair, error) {
	if token == "" {
		return TokenPair{}, ErrInvalidToken
	}
	raw, hash, expiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	userID, err := s.repository.RotateAdminSession(ctx, HashRefreshToken(token), Session{ID: uuid.NewString(), TokenHash: hash, ExpiresAt: expiry})
	if err != nil {
		return TokenPair{}, err
	}
	access, accessExpiry, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: raw, AccessExpiresAt: accessExpiry, RefreshExpiresAt: expiry}, nil
}

func (s *AdminService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidToken
	}
	return s.repository.RevokeAdminSession(ctx, HashRefreshToken(token))
}

func (s *AdminService) ChangePassword(ctx context.Context, userID, current, next string) error {
	hash, err := s.repository.PasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if !security.ComparePassword(hash, current) {
		return ErrInvalidCredentials
	}
	nextHash, err := security.HashPassword(next)
	if err != nil {
		return err
	}
	return s.repository.UpdatePassword(ctx, userID, nextHash)
}
