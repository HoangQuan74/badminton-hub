package permission

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

type Service struct{ repository PermissionRepository }

func NewService(repository PermissionRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	return s.repository.ListPermissions(ctx)
}

func (s *Service) CreatePermission(ctx context.Context, input Input) (model.Permission, error) {
	// Validate and canonicalize the complete permission.
	input, err := normalize(input)
	if err != nil {
		return model.Permission{}, err
	}

	return s.repository.CreatePermission(ctx, input)
}

func (s *Service) GetPermission(ctx context.Context, id string) (model.Permission, error) {
	if _, err := uuid.Parse(id); err != nil {
		return model.Permission{}, identityException.ErrPermissionInvalidInput
	}

	return s.repository.GetPermission(ctx, id)
}

func (s *Service) UpdatePermission(ctx context.Context, id string, input Input) (model.Permission, error) {
	// Validate the identifier and replacement values before persistence.
	if _, err := uuid.Parse(id); err != nil {
		return model.Permission{}, identityException.ErrPermissionInvalidInput
	}

	input, err := normalize(input)
	if err != nil {
		return model.Permission{}, err
	}

	return s.repository.UpdatePermission(
		ctx,
		id,
		input,
	)
}

func (s *Service) DeletePermission(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return identityException.ErrPermissionInvalidInput
	}

	return s.repository.DeletePermission(ctx, id)
}

func normalize(input Input) (Input, error) {
	// Check raw input once, then enforce canonical business values.
	if err := validation.Struct(input); err != nil {
		return Input{}, identityException.ErrPermissionInvalidInput
	}

	input.Module = utils.NormalizeLowercase(input.Module)
	input.Path = strings.TrimSpace(input.Path)
	input.Action = strings.ToUpper(strings.TrimSpace(input.Action))
	input.Description = strings.TrimSpace(input.Description)
	if input.Module == "" || input.Path == "" || strings.ContainsRune(input.Module+input.Path+input.Description, '\x00') {
		return Input{}, identityException.ErrPermissionInvalidInput
	}

	switch input.Action {
	case constant.PermissionActionCreate, constant.PermissionActionRead, constant.PermissionActionUpdate, constant.PermissionActionDelete:
		return input, nil
	default:
		return Input{}, identityException.ErrPermissionInvalidInput
	}
}
