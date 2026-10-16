package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/domain/user"
	"github.com/kainguyen/badminton-hub/be/services/identity/pkg/helper"
	"github.com/kainguyen/badminton-hub/be/services/identity/pkg/utils"
)

type Service struct {
	repo   Repository
	tokens *TokenManager
	now    func() time.Time
}

type RegisterInput struct{ Email, Password, FullName, Phone string }
type LoginInput struct{ Email, Password string }
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

func NewService(repo Repository, tokens *TokenManager) *Service {
	return &Service{repo: repo, tokens: tokens, now: time.Now}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (user.User, TokenPair, error) {
	email, err := utils.NormalizeEmail(input.Email)
	fullName := strings.TrimSpace(input.FullName)
	phone := strings.TrimSpace(input.Phone)
	if err != nil || len(input.Password) < 8 || len(input.Password) > 72 || fullName == "" || len(fullName) > 120 || len(phone) > 20 {
		return user.User{}, TokenPair{}, ErrInvalidInput
	}
	hash, err := helper.HashPassword(input.Password)
	if err != nil {
		return user.User{}, TokenPair{}, fmt.Errorf("hash password: %w", err)
	}
	userID := uuid.NewString()
	access, accessExpiry, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return user.User{}, TokenPair{}, err
	}
	raw, tokenHash, refreshExpiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return user.User{}, TokenPair{}, err
	}
	created, err := s.repo.CreateUserWithSession(ctx,
		CreateUserParams{ID: userID, Email: email, PasswordHash: hash, FullName: fullName, Phone: phone},
		Session{ID: uuid.NewString(), UserID: userID, TokenHash: tokenHash, ExpiresAt: refreshExpiry},
	)
	if err != nil {
		return user.User{}, TokenPair{}, err
	}
	return created, TokenPair{AccessToken: access, RefreshToken: raw, AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (user.User, TokenPair, error) {
	email, err := utils.NormalizeEmail(input.Email)
	if err != nil {
		return user.User{}, TokenPair{}, ErrInvalidCredentials
	}
	found, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil || !helper.ComparePassword(found.PasswordHash, input.Password) {
		return user.User{}, TokenPair{}, ErrInvalidCredentials
	}
	if found.Status != user.StatusActive {
		return user.User{}, TokenPair{}, ErrUserInactive
	}
	pair, err := s.issueSession(ctx, found.ID)
	return found, pair, err
}

func (s *Service) Refresh(ctx context.Context, oldToken string) (TokenPair, error) {
	if oldToken == "" {
		return TokenPair{}, ErrInvalidToken
	}
	raw, hash, refreshExpiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	newSession := Session{ID: uuid.NewString(), TokenHash: hash, ExpiresAt: refreshExpiry}
	owner, err := s.repo.RotateSession(ctx, HashRefreshToken(oldToken), newSession, s.now().UTC())
	if err != nil {
		return TokenPair{}, err
	}
	access, accessExpiry, err := s.tokens.NewAccessToken(owner)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: raw, AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidToken
	}
	return s.repo.RevokeSession(ctx, HashRefreshToken(refreshToken), s.now().UTC())
}

func (s *Service) Me(ctx context.Context, userID string) (user.User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

func (s *Service) issueSession(ctx context.Context, userID string) (TokenPair, error) {
	access, accessExpiry, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return TokenPair{}, err
	}
	raw, hash, refreshExpiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.repo.CreateSession(ctx, Session{ID: uuid.NewString(), UserID: userID, TokenHash: hash, ExpiresAt: refreshExpiry}); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: raw, AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}
