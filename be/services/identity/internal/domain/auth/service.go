package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

type Service struct {
	repo   Repository
	tokens *TokenManager
	now    func() time.Time
}

func NewService(repo Repository, tokens *TokenManager) *Service {
	return &Service{
		repo:   repo,
		tokens: tokens,
		now:    time.Now,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (model.User, TokenPair, error) {
	// Validate the complete registration input.
	if err := validation.Struct(input); err != nil {
		return model.User{}, TokenPair{}, ErrInvalidInput
	}

	// Canonicalize user-provided values before business processing.
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)

	// Hash the password before persistence.
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return model.User{}, TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	// Issue the access token for the new user.
	userID := uuid.NewString()
	access, accessExpiry, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return model.User{}, TokenPair{}, err
	}

	// Issue the refresh token and derive its stored hash.
	raw, tokenHash, refreshExpiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return model.User{}, TokenPair{}, err
	}

	// Prepare the persistence models for the user and session.
	userParams := CreateUserParams{
		ID:           userID,
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		Phone:        input.Phone,
	}
	session := Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: refreshExpiry,
	}

	// Persist the user and session atomically.
	created, err := s.repo.CreateUserWithSession(
		ctx,
		userParams,
		session,
	)
	if err != nil {
		return model.User{}, TokenPair{}, err
	}

	// Build the authentication response tokens.
	tokens := TokenPair{
		AccessToken:      access,
		RefreshToken:     raw,
		AccessExpiresAt:  accessExpiry,
		RefreshExpiresAt: refreshExpiry,
	}

	return created, tokens, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (model.User, TokenPair, error) {
	// Validate the complete login input.
	if err := validation.Struct(input); err != nil {
		return model.User{}, TokenPair{}, ErrInvalidCredentials
	}

	// Canonicalize the email before lookup.
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Load the user and verify the password.
	found, err := s.repo.FindUserByEmail(ctx, input.Email)
	if err != nil || !security.ComparePassword(found.PasswordHash, input.Password) {
		return model.User{}, TokenPair{}, ErrInvalidCredentials
	}

	// Reject inactive accounts before issuing a session.
	if found.Status != model.StatusActive {
		return model.User{}, TokenPair{}, ErrUserInactive
	}

	// Issue a new authenticated session.
	pair, err := s.issueSession(ctx, found.ID)
	return found, pair, err
}

func (s *Service) Refresh(ctx context.Context, oldToken string) (TokenPair, error) {
	// Validate the refresh token before rotating the session.
	if oldToken == "" {
		return TokenPair{}, ErrInvalidToken
	}

	// Generate the replacement refresh token.
	raw, hash, refreshExpiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	// Rotate the persisted session and resolve its owner.
	newSession := Session{
		ID:        uuid.NewString(),
		TokenHash: hash,
		ExpiresAt: refreshExpiry,
	}
	owner, err := s.repo.RotateSession(ctx, HashRefreshToken(oldToken), newSession, s.now().UTC())
	if err != nil {
		return TokenPair{}, err
	}

	// Issue a new access token for the session owner.
	access, accessExpiry, err := s.tokens.NewAccessToken(owner)
	if err != nil {
		return TokenPair{}, err
	}

	// Return the replacement token pair.
	tokens := TokenPair{
		AccessToken:      access,
		RefreshToken:     raw,
		AccessExpiresAt:  accessExpiry,
		RefreshExpiresAt: refreshExpiry,
	}
	return tokens, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	// Validate the refresh token before revoking the session.
	if refreshToken == "" {
		return ErrInvalidToken
	}

	// Revoke the session identified by the hashed refresh token.
	return s.repo.RevokeSession(ctx, HashRefreshToken(refreshToken), s.now().UTC())
}

func (s *Service) Me(ctx context.Context, userID string) (model.User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

func (s *Service) ParseAccessToken(raw string) (string, error) {
	return s.tokens.ParseAccessToken(raw)
}

func (s *Service) issueSession(ctx context.Context, userID string) (TokenPair, error) {
	// Issue the access token for the existing user.
	access, accessExpiry, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return TokenPair{}, err
	}

	// Generate the refresh token and its persisted hash.
	raw, hash, refreshExpiry, err := s.tokens.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	// Persist the new authenticated session.
	session := Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: refreshExpiry,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return TokenPair{}, err
	}

	// Return the issued token pair.
	return TokenPair{
		AccessToken:      access,
		RefreshToken:     raw,
		AccessExpiresAt:  accessExpiry,
		RefreshExpiresAt: refreshExpiry,
	}, nil
}
