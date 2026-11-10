package tenant

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type Repository interface {
	ListPublicVenues(context.Context) ([]model.Venue, error)
	GetPublicVenue(context.Context, string) (model.Venue, error)
	GetPublicVenueByID(context.Context, string) (model.Venue, error)
	ListVenues(context.Context, string) ([]model.Venue, error)
	AuthorizeVenue(context.Context, string, string, bool) (model.Venue, error)
	CreateVenue(context.Context, string, CreateVenueInput) (model.Venue, error)
	UpdateVenue(context.Context, string, string, UpdateVenueInput) (model.Venue, error)
	ArchiveVenue(context.Context, string, string) (model.Venue, error)
	SetOrganizationStaff(context.Context, string, SetOrganizationStaffInput) (model.OrganizationStaff, error)
	SetVenueStaff(context.Context, string, SetVenueStaffInput) (model.VenueStaff, error)
	ListOrganizations(context.Context, string) ([]model.Organization, error)
	GetOrganization(context.Context, string, string) (model.Organization, error)
	CreateOrganization(context.Context, string, CreateOrganizationInput) (model.Organization, error)
	UpdateOrganization(context.Context, string, string, UpdateOrganizationInput) (model.Organization, error)
	ArchiveOrganization(context.Context, string, string) (model.Organization, error)
}
