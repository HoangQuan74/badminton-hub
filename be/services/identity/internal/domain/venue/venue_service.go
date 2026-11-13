package venue

import (
	"context"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/constant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

func (s *Service) ListPublicVenues(ctx context.Context) ([]model.Venue, error) {
	return s.repository.ListPublicVenues(ctx)
}

func (s *Service) GetPublicVenue(ctx context.Context, slug string) (model.Venue, error) {
	slug = utils.NormalizeLowercase(slug)
	if !utils.IsSlug(slug, 100) {
		return model.Venue{}, identityException.ErrTenantInvalidInput
	}

	return s.repository.GetPublicVenue(ctx, slug)
}

func (s *Service) GetPublicVenueByID(ctx context.Context, id string) (model.Venue, error) {
	if strings.TrimSpace(id) == "" {
		return model.Venue{}, identityException.ErrTenantInvalidInput
	}

	return s.repository.GetPublicVenueByID(ctx, id)
}

func (s *Service) ListVenues(ctx context.Context, actorID string) ([]model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, identityException.ErrTenantForbidden
	}

	return s.repository.ListVenues(ctx, actorID)
}

func (s *Service) AuthorizeVenue(ctx context.Context, actorID, venueID string, requireOwner bool) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(venueID) == "" {
		return model.Venue{}, identityException.ErrTenantInvalidInput
	}

	return s.repository.AuthorizeVenue(
		ctx,
		actorID,
		venueID,
		requireOwner,
	)
}

func (s *Service) CreateVenue(ctx context.Context, actorID string, input CreateVenueInput) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, identityException.ErrTenantForbidden
	}

	if err := validation.Struct(input); err != nil {
		return model.Venue{}, identityException.ErrTenantInvalidInput
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Slug = utils.NormalizeLowercase(input.Slug)
	input.Address = strings.TrimSpace(input.Address)
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.Timezone == "" {
		input.Timezone = "Asia/Ho_Chi_Minh"
	}

	if !utils.IsRequiredText(input.Name, 120) || !utils.IsSlug(input.Slug, 100) || utf8.RuneCountInString(input.Address) > 500 || !validTimezone(input.Timezone) {
		return model.Venue{}, identityException.ErrTenantInvalidInput
	}

	// Allow new venues only for an owner of an approved organization.
	access, err := s.repository.GetOrganizationAccess(
		ctx,
		actorID,
		input.OrganizationID,
	)
	if err != nil {
		return model.Venue{}, err
	}

	if access.Role != constant.RoleTenantOwner || access.Status != constant.OrganizationStatusActive {
		return model.Venue{}, identityException.ErrTenantForbidden
	}

	return s.repository.CreateVenue(
		ctx,
		actorID,
		input,
	)
}

func (s *Service) UpdateVenue(ctx context.Context, actorID, venueID string, input UpdateVenueInput) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(venueID) == "" || (input.Name == nil && input.Slug == nil && input.Address == nil && input.Timezone == nil) {
		return model.Venue{}, identityException.ErrTenantInvalidInput
	}

	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !utils.IsRequiredText(value, 120) {
			return model.Venue{}, identityException.ErrTenantInvalidInput
		}

		input.Name = &value
	}

	if input.Slug != nil {
		value := utils.NormalizeLowercase(*input.Slug)
		if !utils.IsSlug(value, 100) {
			return model.Venue{}, identityException.ErrTenantInvalidInput
		}

		input.Slug = &value
	}

	if input.Address != nil {
		value := strings.TrimSpace(*input.Address)
		if utf8.RuneCountInString(value) > 500 {
			return model.Venue{}, identityException.ErrTenantInvalidInput
		}

		input.Address = &value
	}

	if input.Timezone != nil {
		value := strings.TrimSpace(*input.Timezone)
		if !validTimezone(value) {
			return model.Venue{}, identityException.ErrTenantInvalidInput
		}

		input.Timezone = &value
	}

	// Authorize the current owner before persisting venue changes.
	if _, err := s.repository.AuthorizeVenue(
		ctx,
		actorID,
		venueID,
		true,
	); err != nil {
		return model.Venue{}, err
	}

	return s.repository.UpdateVenue(
		ctx,
		actorID,
		venueID,
		input,
	)
}

func (s *Service) ArchiveVenue(ctx context.Context, actorID, venueID string) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, identityException.ErrTenantForbidden
	}

	if strings.TrimSpace(venueID) == "" {
		return model.Venue{}, identityException.ErrTenantInvalidInput
	}

	// Authorize the current owner before archival.
	if _, err := s.repository.AuthorizeVenue(
		ctx,
		actorID,
		venueID,
		true,
	); err != nil {
		return model.Venue{}, err
	}

	return s.repository.ArchiveVenue(
		ctx,
		actorID,
		venueID,
	)
}

func validTimezone(value string) bool {
	_, err := time.LoadLocation(value)
	return err == nil && value != "Local"
}

type Service struct{ repository VenueRepository }

func NewService(repository VenueRepository) *Service {
	return &Service{repository: repository}
}
