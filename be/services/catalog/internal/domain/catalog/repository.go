package catalog

import (
	"context"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

type Repository interface {
	ListOrganizations(context.Context, string) ([]model.Organization, error)
	GetOrganization(context.Context, string, string) (model.Organization, error)
	CreateOrganization(context.Context, string, CreateOrganizationInput) (model.Organization, error)
	UpdateOrganization(context.Context, string, string, UpdateOrganizationInput) (model.Organization, error)
	ArchiveOrganization(context.Context, string, string) (model.Organization, error)
	GetPublicVenue(context.Context, string) (model.Venue, error)
	GetTenantVenue(context.Context, string, string) (model.Venue, error)
	CreateVenue(context.Context, string, CreateVenueInput) (model.Venue, error)
	UpdateVenue(context.Context, string, string, UpdateVenueInput) (model.Venue, error)
	ArchiveVenue(context.Context, string, string) (model.Venue, error)
	GetTenantCourt(context.Context, string, string) (model.Court, error)
	ListPublicVenues(context.Context) ([]model.Venue, error)
	ListPublicVenueCourts(context.Context, string) ([]model.Court, error)
	ListTenantVenues(context.Context, string) ([]model.Venue, error)
	ListTenantCourts(context.Context, string, string) ([]model.Court, error)
	CreateCourt(context.Context, string, string, string, int32) (model.Court, error)
	UpdateCourt(context.Context, string, string, string, int32, string) (model.Court, error)
	ArchiveCourt(context.Context, string, string) (model.Court, error)
}
