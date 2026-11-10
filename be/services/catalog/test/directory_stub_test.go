package test

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	"google.golang.org/grpc/metadata"
)

// directoryStub represents the Identity boundary in Catalog-only tests.
type directoryStub struct{ venues map[string]model.Venue }

func newDirectoryStub() *directoryStub {
	return &directoryStub{venues: map[string]model.Venue{
		"20000000-0000-4000-8000-000000000001": {ID: "20000000-0000-4000-8000-000000000001", OrganizationID: "10000000-0000-4000-8000-000000000001", Status: "PUBLISHED"},
	}}
}
func (*directoryStub) ListOrganizations(context.Context) ([]model.Organization, error) {
	return nil, nil
}
func (*directoryStub) GetOrganization(context.Context, string) (model.Organization, error) {
	return model.Organization{}, nil
}
func (*directoryStub) CreateOrganization(context.Context, catalog.CreateOrganizationInput) (model.Organization, error) {
	return model.Organization{}, nil
}
func (*directoryStub) UpdateOrganization(context.Context, string, catalog.UpdateOrganizationInput) (model.Organization, error) {
	return model.Organization{}, nil
}
func (*directoryStub) ArchiveOrganization(context.Context, string) (model.Organization, error) {
	return model.Organization{}, nil
}
func (d *directoryStub) ListPublicVenues(context.Context) ([]model.Venue, error) { return nil, nil }
func (d *directoryStub) GetPublicVenue(context.Context, string) (model.Venue, error) {
	return model.Venue{}, nil
}
func (d *directoryStub) GetPublicVenueByID(_ context.Context, id string) (model.Venue, error) {
	venue, ok := d.venues[id]
	if !ok || venue.Status != "PUBLISHED" {
		return model.Venue{}, catalog.ErrNotFound
	}
	return venue, nil
}
func (d *directoryStub) ListVenues(context.Context) ([]model.Venue, error) { return nil, nil }
func (d *directoryStub) AuthorizeVenue(ctx context.Context, id string, ownerOnly bool) (model.Venue, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	auth := strings.Join(md.Get("authorization"), "")
	venue, ok := d.venues[id]
	if !ok || venue.Status == "ARCHIVED" || auth == "Bearer outsider" || (ownerOnly && auth != "Bearer owner") {
		return model.Venue{}, catalog.ErrForbidden
	}
	return venue, nil
}
func (d *directoryStub) CreateVenue(_ context.Context, input catalog.CreateVenueInput) (model.Venue, error) {
	item := model.Venue{ID: uuid.NewString(), OrganizationID: input.OrganizationID, Name: input.Name, Slug: input.Slug, Status: "DRAFT"}
	d.venues[item.ID] = item
	return item, nil
}
func (d *directoryStub) UpdateVenue(context.Context, string, catalog.UpdateVenueInput) (model.Venue, error) {
	return model.Venue{}, nil
}
func (d *directoryStub) ArchiveVenue(_ context.Context, id string) (model.Venue, error) {
	item := d.venues[id]
	item.Status = "ARCHIVED"
	d.venues[id] = item
	return item, nil
}
