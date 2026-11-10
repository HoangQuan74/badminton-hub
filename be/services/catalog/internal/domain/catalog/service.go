package catalog

import (
	"context"
	"strings"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/validation"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) ListPublicVenues(ctx context.Context) ([]model.Venue, error) {
	return s.repository.ListPublicVenues(ctx)
}

func (s *Service) ListPublicVenueCourts(ctx context.Context, venueID string) ([]model.Court, error) {
	if strings.TrimSpace(venueID) == "" {
		return nil, ErrInvalidInput
	}
	return s.repository.ListPublicVenueCourts(ctx, venueID)
}

func (s *Service) ListTenantVenues(ctx context.Context, actorID string) ([]model.Venue, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, ErrForbidden
	}
	return s.repository.ListTenantVenues(ctx, actorID)
}

func (s *Service) ListTenantCourts(ctx context.Context, actorID, venueID string) ([]model.Court, error) {
	if strings.TrimSpace(actorID) == "" {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(venueID) == "" {
		return nil, ErrInvalidInput
	}
	return s.repository.ListTenantCourts(ctx, actorID, venueID)
}

func (s *Service) CreateCourt(ctx context.Context, actorID, venueID, name string, interval int32) (model.Court, error) {
	if strings.TrimSpace(actorID) == "" || strings.TrimSpace(venueID) == "" {
		return model.Court{}, ErrForbidden
	}
	// Validate the input before canonicalizing user-provided values.
	input := courtInput{Name: name, SlotIntervalMinutes: interval}
	if err := validation.Struct(input); err != nil {
		return model.Court{}, ErrInvalidInput
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 80 {
		return model.Court{}, ErrInvalidInput
	}
	if interval == 0 {
		interval = 30
	}
	if interval < 15 || interval > 180 || interval%15 != 0 {
		return model.Court{}, ErrInvalidInput
	}
	return s.repository.CreateCourt(ctx, actorID, venueID, name, interval)
}

func (s *Service) UpdateCourt(ctx context.Context, actorID, courtID, name string, interval int32, status string) (model.Court, error) {
	if strings.TrimSpace(actorID) == "" || strings.TrimSpace(courtID) == "" {
		return model.Court{}, ErrForbidden
	}
	// Validate the input before canonicalizing user-provided values.
	input := courtInput{Name: name, SlotIntervalMinutes: interval, Status: status}
	if err := validation.Struct(input); err != nil {
		return model.Court{}, ErrInvalidInput
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 80 {
		return model.Court{}, ErrInvalidInput
	}
	if interval < 15 || interval > 180 || interval%15 != 0 {
		return model.Court{}, ErrInvalidInput
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "ACTIVE" && status != "INACTIVE" {
		return model.Court{}, ErrUnsupportedStatus
	}
	return s.repository.UpdateCourt(ctx, actorID, courtID, name, interval, status)
}

func (s *Service) ArchiveCourt(ctx context.Context, actorID, courtID string) (model.Court, error) {
	if strings.TrimSpace(actorID) == "" || strings.TrimSpace(courtID) == "" {
		return model.Court{}, ErrForbidden
	}
	return s.repository.ArchiveCourt(ctx, actorID, courtID)
}
