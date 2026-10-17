package staff

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

func (s *Service) List(ctx context.Context, query model.ListQuery) ([]model.Staff, error) {
	// Execute the shared list query through the staff repository.
	return s.repository.List(ctx, query)
}

func (s *Service) Get(ctx context.Context, id string) (model.Staff, error) {
	// Validate the staff identifier before loading the record.
	if _, err := uuid.Parse(id); err != nil {
		return model.Staff{}, ErrInvalidInput
	}
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (model.Staff, error) {
	// Validate the complete staff creation input.
	if err := validation.Struct(input); err != nil {
		return model.Staff{}, ErrInvalidInput
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.RoleID = strings.TrimSpace(input.RoleID)

	// Hash the password before persistence.
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return model.Staff{}, err
	}

	// Prepare the persistence input for the staff record.
	params := CreateParams{
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		RoleID:       input.RoleID,
		IsAdmin:      input.IsAdmin,
	}
	return s.repository.Create(ctx, params)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (model.Staff, error) {
	// Validate the staff identifier before updating the record.
	if _, err := uuid.Parse(input.ID); err != nil {
		return model.Staff{}, ErrInvalidInput
	}
	if err := validation.Struct(input); err != nil {
		return model.Staff{}, ErrInvalidInput
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.Status = strings.TrimSpace(input.Status)

	// Persist the canonicalized staff fields.
	return s.repository.Update(ctx, input)
}

func (s *Service) UpdateAccess(ctx context.Context, input AccessInput) (model.Staff, error) {
	// Validate the staff identifier before changing access.
	if _, err := uuid.Parse(input.ID); err != nil {
		return model.Staff{}, ErrInvalidInput
	}
	input.RoleID = strings.TrimSpace(input.RoleID)

	// Persist the updated access settings.
	return s.repository.UpdateAccess(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	// Validate the staff identifier before deleting the record.
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidInput
	}
	return s.repository.Delete(ctx, id)
}
