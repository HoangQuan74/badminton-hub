package user

import (
	"context"
	"strings"

	"github.com/google/uuid"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, query model.ListQuery) ([]model.User, error) {
	// Execute the shared list query through the user repository.
	return s.repository.List(ctx, query)
}

func (s *Service) Get(ctx context.Context, id string) (model.User, error) {
	// Validate the user identifier before loading the record.
	if _, err := uuid.Parse(id); err != nil {
		return model.User{}, ErrInvalidInput
	}
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (model.User, error) {
	// Validate the complete user creation input.
	if err := validation.Struct(input); err != nil {
		return model.User{}, ErrInvalidInput
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
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
		Status:       "active",
	}
	return s.repository.Create(ctx, params)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (model.User, error) {
	// Validate the user identifier before updating the record.
	if _, err := uuid.Parse(input.ID); err != nil {
		return model.User{}, ErrInvalidInput
	}
	if err := validation.Struct(input); err != nil {
		return model.User{}, ErrInvalidInput
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Status = strings.TrimSpace(input.Status)

	// Persist the canonicalized user fields.
	return s.repository.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	// Validate the user identifier before deleting the record.
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidInput
	}
	return s.repository.Delete(ctx, id)
}
