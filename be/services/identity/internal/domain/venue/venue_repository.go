package venue

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type VenueRepository interface {
	GetOrganizationAccess(ctx context.Context, actorID, organizationID string) (model.VenueOrganizationAccess, error)
	ListPublicVenues(ctx context.Context) ([]model.Venue, error)
	GetPublicVenue(ctx context.Context, slug string) (model.Venue, error)
	GetPublicVenueByID(ctx context.Context, venueID string) (model.Venue, error)
	ListVenues(ctx context.Context, actorID string) ([]model.Venue, error)
	AuthorizeVenue(ctx context.Context, actorID, venueID string, requireOwner bool) (model.Venue, error)
	CreateVenue(ctx context.Context, actorID string, input CreateVenueInput) (model.Venue, error)
	UpdateVenue(ctx context.Context, actorID, venueID string, input UpdateVenueInput) (model.Venue, error)
	ArchiveVenue(ctx context.Context, actorID, venueID string) (model.Venue, error)
}
