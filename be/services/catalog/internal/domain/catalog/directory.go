package catalog

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

// Directory is implemented by Identity, the authority for tenant and venue state.
type Directory interface {
	ListPublicVenues(context.Context) ([]model.Venue, error)
	GetPublicVenue(context.Context, string) (model.Venue, error)
	GetPublicVenueByID(context.Context, string) (model.Venue, error)
	ListVenues(context.Context) ([]model.Venue, error)
	AuthorizeVenue(context.Context, string, bool) (model.Venue, error)
	CreateVenue(context.Context, CreateVenueInput) (model.Venue, error)
	UpdateVenue(context.Context, string, UpdateVenueInput) (model.Venue, error)
	ArchiveVenue(context.Context, string) (model.Venue, error)
}
