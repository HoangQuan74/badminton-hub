package staff

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
	repository StaffRepository
}

func NewService(repository StaffRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListStaff(ctx context.Context, query model.ListQuery) ([]model.Staff, error) {
	// Execute the shared list query through the staff repository.
	return s.repository.ListStaff(ctx, query)
}

func (s *Service) GetStaff(ctx context.Context, id string) (model.Staff, error) {
	// Validate the staff identifier before loading the record.
	if _, err := uuid.Parse(id); err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	return s.repository.GetStaff(ctx, id)
}

func (s *Service) CreateStaff(ctx context.Context, input CreateInput) (model.Staff, error) {
	// Validate the complete staff creation input.
	if err := validation.Struct(input); err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	input.Email = utils.NormalizeLowercase(input.Email)
	input.FullName = strings.TrimSpace(input.FullName)
	input.RoleID = strings.TrimSpace(input.RoleID)

	// Hash the password before persistence.
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return model.Staff{}, err
	}

	// Prepare the persistence input for the staff record.
	params := CreateStaffParams{
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		RoleID:       input.RoleID,
		IsAdmin:      input.IsAdmin,
	}
	return s.repository.CreateStaff(ctx, params)
}

func (s *Service) UpdateStaff(ctx context.Context, input UpdateStaffInput) (model.Staff, error) {
	// Validate the staff identifier before updating the record.
	if _, err := uuid.Parse(input.ID); err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	if err := validation.Struct(input); err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	input.Email = utils.NormalizeLowercase(input.Email)
	input.FullName = strings.TrimSpace(input.FullName)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status != constant.AccountStatusActive && input.Status != constant.AccountStatusSuspended {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	// Persist the canonicalized staff fields.
	return s.repository.UpdateStaff(ctx, input)
}

func (s *Service) UpdateStaffAccess(ctx context.Context, input UpdateStaffAccessInput) (model.Staff, error) {
	// Validate the staff identifier before changing access.
	if _, err := uuid.Parse(input.ID); err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	input.RoleID = strings.TrimSpace(input.RoleID)

	// Persist the updated access settings.
	return s.repository.UpdateStaffAccess(ctx, input)
}

func (s *Service) DeleteStaff(ctx context.Context, id string) error {
	// Validate the staff identifier before deleting the record.
	if _, err := uuid.Parse(id); err != nil {
		return identityException.ErrStaffInvalidInput
	}

	return s.repository.DeleteStaff(ctx, id)
}
