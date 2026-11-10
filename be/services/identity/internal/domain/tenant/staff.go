package tenant

import (
	"context"
	"strings"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

func (s *Service) SetOrganizationStaff(ctx context.Context, actorID string, input SetOrganizationStaffInput) (model.OrganizationStaff, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.OrganizationStaff{}, ErrForbidden
	}
	if strings.TrimSpace(input.OrganizationID) == "" || strings.TrimSpace(input.UserID) == "" ||
		(input.Role != "tenant_staff" && input.Role != "venue_staff") ||
		(input.Status != "ACTIVE" && input.Status != "INACTIVE") {
		return model.OrganizationStaff{}, ErrInvalidInput
	}
	return s.repository.SetOrganizationStaff(ctx, actorID, input)
}

func (s *Service) SetVenueStaff(ctx context.Context, actorID string, input SetVenueStaffInput) (model.VenueStaff, error) {
	if strings.TrimSpace(actorID) == "" {
		return model.VenueStaff{}, ErrForbidden
	}
	if strings.TrimSpace(input.VenueID) == "" || strings.TrimSpace(input.UserID) == "" ||
		(input.Status != "ACTIVE" && input.Status != "INACTIVE") {
		return model.VenueStaff{}, ErrInvalidInput
	}
	return s.repository.SetVenueStaff(ctx, actorID, input)
}
