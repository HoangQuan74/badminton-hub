package catalog

import (
	"context"
	"strings"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/validation"
)

func (s *Service) ListOrganizations(ctx context.Context, actorID string) ([]model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, ErrForbidden
	}
	return s.repository.ListOrganizations(ctx, actorID)
}

func (s *Service) GetOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, ErrForbidden
	}
	if strings.TrimSpace(organizationID) == "" {
		return model.Organization{}, ErrInvalidInput
	}
	return s.repository.GetOrganization(ctx, actorID, organizationID)
}

func (s *Service) CreateOrganization(ctx context.Context, actorID string, input CreateOrganizationInput) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, ErrForbidden
	}
	if err := validation.Struct(input); err != nil {
		return model.Organization{}, ErrInvalidInput
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	if !validCatalogName(input.Name) || !validCatalogSlug(input.Slug) {
		return model.Organization{}, ErrInvalidInput
	}
	return s.repository.CreateOrganization(ctx, actorID, input)
}

func (s *Service) UpdateOrganization(ctx context.Context, actorID, organizationID string, input UpdateOrganizationInput) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, ErrForbidden
	}
	if strings.TrimSpace(organizationID) == "" || (input.Name == nil && input.Slug == nil) {
		return model.Organization{}, ErrInvalidInput
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !validCatalogName(value) {
			return model.Organization{}, ErrInvalidInput
		}
		input.Name = &value
	}
	if input.Slug != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Slug))
		if !validCatalogSlug(value) {
			return model.Organization{}, ErrInvalidInput
		}
		input.Slug = &value
	}
	return s.repository.UpdateOrganization(ctx, actorID, organizationID, input)
}

func (s *Service) ArchiveOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, ErrForbidden
	}
	if strings.TrimSpace(organizationID) == "" {
		return model.Organization{}, ErrInvalidInput
	}
	return s.repository.ArchiveOrganization(ctx, actorID, organizationID)
}
