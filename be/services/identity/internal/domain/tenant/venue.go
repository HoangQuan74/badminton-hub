package tenant

import (
	"context"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/validation"
)

func (s *Service) ListPublicVenues(ctx context.Context) ([]model.Venue, error) {
	return s.repository.ListPublicVenues(ctx)
}
func (s *Service) GetPublicVenue(ctx context.Context, slug string) (model.Venue, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !validSlug(slug) {
		return model.Venue{}, ErrInvalidInput
	}
	return s.repository.GetPublicVenue(ctx, slug)
}
func (s *Service) GetPublicVenueByID(ctx context.Context, id string) (model.Venue, error) {
	if strings.TrimSpace(id) == "" {
		return model.Venue{}, ErrInvalidInput
	}
	return s.repository.GetPublicVenueByID(ctx, id)
}
func (s *Service) ListVenues(ctx context.Context, actorID string) ([]model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, ErrForbidden
	}
	return s.repository.ListVenues(ctx, actorID)
}
func (s *Service) AuthorizeVenue(ctx context.Context, actorID, venueID string, requireOwner bool) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, ErrForbidden
	}
	if strings.TrimSpace(venueID) == "" {
		return model.Venue{}, ErrInvalidInput
	}
	return s.repository.AuthorizeVenue(ctx, actorID, venueID, requireOwner)
}
func (s *Service) CreateVenue(ctx context.Context, actorID string, input CreateVenueInput) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, ErrForbidden
	}
	if err := validation.Struct(input); err != nil {
		return model.Venue{}, ErrInvalidInput
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Address = strings.TrimSpace(input.Address)
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.Timezone == "" {
		input.Timezone = "Asia/Ho_Chi_Minh"
	}
	if !validName(input.Name) || !validSlug(input.Slug) || utf8.RuneCountInString(input.Address) > 500 || !validTimezone(input.Timezone) {
		return model.Venue{}, ErrInvalidInput
	}
	return s.repository.CreateVenue(ctx, actorID, input)
}
func (s *Service) UpdateVenue(ctx context.Context, actorID, venueID string, input UpdateVenueInput) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, ErrForbidden
	}
	if strings.TrimSpace(venueID) == "" || (input.Name == nil && input.Slug == nil && input.Address == nil && input.Timezone == nil) {
		return model.Venue{}, ErrInvalidInput
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !validName(value) {
			return model.Venue{}, ErrInvalidInput
		}
		input.Name = &value
	}
	if input.Slug != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Slug))
		if !validSlug(value) {
			return model.Venue{}, ErrInvalidInput
		}
		input.Slug = &value
	}
	if input.Address != nil {
		value := strings.TrimSpace(*input.Address)
		if utf8.RuneCountInString(value) > 500 {
			return model.Venue{}, ErrInvalidInput
		}
		input.Address = &value
	}
	if input.Timezone != nil {
		value := strings.TrimSpace(*input.Timezone)
		if !validTimezone(value) {
			return model.Venue{}, ErrInvalidInput
		}
		input.Timezone = &value
	}
	return s.repository.UpdateVenue(ctx, actorID, venueID, input)
}
func (s *Service) ArchiveVenue(ctx context.Context, actorID, venueID string) (model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.Venue{}, ErrForbidden
	}
	if strings.TrimSpace(venueID) == "" {
		return model.Venue{}, ErrInvalidInput
	}
	return s.repository.ArchiveVenue(ctx, actorID, venueID)
}
func validTimezone(value string) bool {
	_, err := time.LoadLocation(value)
	return err == nil && value != "Local"
}
