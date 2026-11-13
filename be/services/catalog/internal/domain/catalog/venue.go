package catalog

import (
	"context"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"

	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/validation"
)

var catalogSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *Service) GetPublicVenue(ctx context.Context, slug string) (model.Venue, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !validCatalogSlug(slug) {
		return model.Venue{}, catalogException.ErrInvalidInput
	}

	return s.repository.GetPublicVenue(ctx, slug)
}

func (s *Service) GetTenantVenue(ctx context.Context, actorID, venueID string) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, catalogException.ErrForbidden
	}

	if strings.TrimSpace(venueID) == "" {
		return model.Venue{}, catalogException.ErrInvalidInput
	}

	return s.repository.GetTenantVenue(
		ctx,
		actorID,
		venueID,
	)
}

func (s *Service) GetTenantCourt(ctx context.Context, actorID, courtID string) (model.Court, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Court{}, catalogException.ErrForbidden
	}

	if strings.TrimSpace(courtID) == "" {
		return model.Court{}, catalogException.ErrInvalidInput
	}

	return s.repository.GetTenantCourt(
		ctx,
		actorID,
		courtID,
	)
}

func (s *Service) CreateVenue(ctx context.Context, actorID string, input CreateVenueInput) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, catalogException.ErrForbidden
	}

	if err := validation.Struct(input); err != nil {
		return model.Venue{}, catalogException.ErrInvalidInput
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Address = strings.TrimSpace(input.Address)
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.Timezone == "" {
		input.Timezone = "Asia/Ho_Chi_Minh"
	}

	if !validCatalogName(input.Name) || !validCatalogSlug(input.Slug) || utf8.RuneCountInString(input.Address) > 500 || !validTimezone(input.Timezone) {
		return model.Venue{}, catalogException.ErrInvalidInput
	}

	return s.repository.CreateVenue(
		ctx,
		actorID,
		input,
	)
}

func (s *Service) UpdateVenue(ctx context.Context, actorID, venueID string, input UpdateVenueInput) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, catalogException.ErrForbidden
	}

	if strings.TrimSpace(venueID) == "" || (input.Name == nil && input.Slug == nil && input.Address == nil && input.Timezone == nil) {
		return model.Venue{}, catalogException.ErrInvalidInput
	}

	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !validCatalogName(value) {
			return model.Venue{}, catalogException.ErrInvalidInput
		}

		input.Name = &value
	}

	if input.Slug != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Slug))
		if !validCatalogSlug(value) {
			return model.Venue{}, catalogException.ErrInvalidInput
		}

		input.Slug = &value
	}

	if input.Address != nil {
		value := strings.TrimSpace(*input.Address)
		if utf8.RuneCountInString(value) > 500 {
			return model.Venue{}, catalogException.ErrInvalidInput
		}

		input.Address = &value
	}

	if input.Timezone != nil {
		value := strings.TrimSpace(*input.Timezone)
		if !validTimezone(value) {
			return model.Venue{}, catalogException.ErrInvalidInput
		}

		input.Timezone = &value
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
		return model.Venue{}, catalogException.ErrForbidden
	}

	if strings.TrimSpace(venueID) == "" {
		return model.Venue{}, catalogException.ErrInvalidInput
	}

	return s.repository.ArchiveVenue(
		ctx,
		actorID,
		venueID,
	)
}

func validCatalogName(value string) bool {
	return value != "" && utf8.RuneCountInString(value) <= 120
}

func validCatalogSlug(value string) bool {
	return len(value) <= 100 && catalogSlugPattern.MatchString(value)
}

func validTimezone(value string) bool {
	_, err := time.LoadLocation(value)
	return err == nil && value != "Local"
}
