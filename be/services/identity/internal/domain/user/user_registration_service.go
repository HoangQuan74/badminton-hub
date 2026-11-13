package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

type RegistrationService struct {
	repo   RegistrationRepository
	tokens *auth.TokenManager
}

func NewRegistrationService(repository RegistrationRepository, tokens *auth.TokenManager) *RegistrationService {
	return &RegistrationService{
		repo:   repository,
		tokens: tokens,
	}
}

func (s *RegistrationService) RegisterUser(ctx context.Context, input RegisterInput) (model.User, model.TokenPair, error) {
	// Validate the complete registration input.
	if err := validation.Struct(input); err != nil {
		return model.User{}, model.TokenPair{}, identityException.ErrAuthInvalidInput
	}

	// Canonicalize user-provided values before business processing.
	input.Email = utils.NormalizeLowercase(input.Email)
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)

	// Hash the password before persistence.
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	// Prepare the new account's tokens and hashed session together.
	userID := uuid.NewString()
	tokens, session, err := s.tokens.NewSession(userID)
	if err != nil {
		return model.User{}, model.TokenPair{}, err
	}

	// Prepare the persistence models for the user and session.
	userParams := RegistrationParams{
		ID:           userID,
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		Phone:        input.Phone,
	}

	// Persist the user and session atomically.
	created, err := s.repo.CreateUserWithSession(
		ctx,
		userParams,
		session,
	)
	if err != nil {
		return model.User{}, model.TokenPair{}, err
	}

	return created, tokens, nil
}
