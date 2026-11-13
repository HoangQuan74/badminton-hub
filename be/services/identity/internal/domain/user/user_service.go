package user

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

type Service struct {
	repository UserRepository
}

func NewService(repository UserRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListUsers(ctx context.Context, query model.ListQuery) ([]model.User, error) {
	// Execute the shared list query through the user repository.
	return s.repository.ListUsers(ctx, query)
}

func (s *Service) GetUser(ctx context.Context, id string) (model.User, error) {
	// Validate the user identifier before loading the record.
	if _, err := uuid.Parse(id); err != nil {
		return model.User{}, identityException.ErrUserInvalidInput
	}

	return s.repository.GetUser(ctx, id)
}

func (s *Service) CreateUser(ctx context.Context, input CreateInput) (model.User, error) {
	// Validate the complete user creation input.
	if err := validation.Struct(input); err != nil {
		return model.User{}, identityException.ErrUserInvalidInput
	}

	input.Email = utils.NormalizeLowercase(input.Email)
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)

	// Hash the password before persistence.
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return model.User{}, err
	}

	// Prepare the persistence input for an active user.
	params := CreateParams{
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		Phone:        input.Phone,
		Status:       constant.AccountStatusActive,
	}
	return s.repository.CreateUser(ctx, params)
}

func (s *Service) UpdateUser(ctx context.Context, input UpdateInput) (model.User, error) {
	// Validate the user identifier before updating the record.
	if _, err := uuid.Parse(input.ID); err != nil {
		return model.User{}, identityException.ErrUserInvalidInput
	}

	if err := validation.Struct(input); err != nil {
		return model.User{}, identityException.ErrUserInvalidInput
	}

	input.Email = utils.NormalizeLowercase(input.Email)
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status != constant.AccountStatusActive && input.Status != constant.AccountStatusSuspended {
		return model.User{}, identityException.ErrUserInvalidInput
	}

	// Persist the canonicalized user fields.
	return s.repository.UpdateUser(ctx, input)
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	// Validate the user identifier before deleting the record.
	if _, err := uuid.Parse(id); err != nil {
		return identityException.ErrUserInvalidInput
	}

	return s.repository.DeleteUser(ctx, id)
}

func CanAccessPortal(user model.User) bool {
	return !user.IsAdmin
}
