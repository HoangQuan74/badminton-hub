package catalog

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

// Directory is implemented by Identity, the authority for tenant and venue state.
type Directory interface {
	ListOrganizations(context.Context) ([]model.Organization, error)
	GetOrganization(context.Context, string) (model.Organization, error)
	CreateOrganization(context.Context, CreateOrganizationInput) (model.Organization, error)
	UpdateOrganization(context.Context, string, UpdateOrganizationInput) (model.Organization, error)
	ArchiveOrganization(context.Context, string) (model.Organization, error)
	ListPublicVenues(context.Context) ([]model.Venue, error)
	GetPublicVenue(context.Context, string) (model.Venue, error)
	GetPublicVenueByID(context.Context, string) (model.Venue, error)
	ListVenues(context.Context) ([]model.Venue, error)
	AuthorizeVenue(context.Context, string, bool) (model.Venue, error)
	CreateVenue(context.Context, CreateVenueInput) (model.Venue, error)
	UpdateVenue(context.Context, string, UpdateVenueInput) (model.Venue, error)
	ArchiveVenue(context.Context, string) (model.Venue, error)
}
