package tenant

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) ListOrganizations(ctx context.Context, actorID string) ([]model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, ErrForbidden
	}
	return s.repository.ListOrganizations(ctx, actorID)
}

func (s *Service) GetOrganization(ctx context.Context, actorID, id string) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, ErrForbidden
	}
	if strings.TrimSpace(id) == "" {
		return model.Organization{}, ErrInvalidInput
	}
	return s.repository.GetOrganization(ctx, actorID, id)
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
	if !validName(input.Name) || !validSlug(input.Slug) {
		return model.Organization{}, ErrInvalidInput
	}
	return s.repository.CreateOrganization(ctx, actorID, input)
}

func (s *Service) UpdateOrganization(ctx context.Context, actorID, id string, input UpdateOrganizationInput) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, ErrForbidden
	}
	if strings.TrimSpace(id) == "" || (input.Name == nil && input.Slug == nil) {
		return model.Organization{}, ErrInvalidInput
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !validName(value) {
			return model.Organization{}, ErrInvalidInput
		}
		input.Name = &value
	}
	if input.Slug != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Slug))
		if !validSlug(value) {
			return model.Organization{}, ErrInvalidInput
		}
		input.Slug = &value
	}
	return s.repository.UpdateOrganization(ctx, actorID, id, input)
}

func (s *Service) ArchiveOrganization(ctx context.Context, actorID, id string) (model.Organization, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Organization{}, ErrForbidden
	}
	if strings.TrimSpace(id) == "" {
		return model.Organization{}, ErrInvalidInput
	}
	return s.repository.ArchiveOrganization(ctx, actorID, id)
}

func validName(value string) bool { return value != "" && utf8.RuneCountInString(value) <= 120 }
func validSlug(value string) bool { return len(value) <= 100 && slugPattern.MatchString(value) }
