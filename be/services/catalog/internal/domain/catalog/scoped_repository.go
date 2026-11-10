package catalog

import (
	"context"
	"errors"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

// Store contains only Catalog-owned court and image persistence operations.
type Store interface {
	ListCourts(context.Context, string, string, bool) ([]model.Court, error)
	GetCourt(context.Context, string) (model.Court, error)
	CreateCourt(context.Context, string, string, string, int32) (model.Court, error)
	UpdateCourt(context.Context, model.Court, string, int32, string) (model.Court, error)
	ArchiveCourt(context.Context, model.Court) (model.Court, error)
	SaveImage(context.Context, model.VenueImage) (model.VenueImage, error)
	ListImages(context.Context, string) ([]model.VenueImage, error)
	GetImage(context.Context, string) (model.VenueImage, error)
}

type ScopedRepository struct {
	store     Store
	directory Directory
}

func NewScopedRepository(store Store, directory Directory) *ScopedRepository {
	return &ScopedRepository{store: store, directory: directory}
}

func (r *ScopedRepository) ListOrganizations(ctx context.Context, _ string) ([]model.Organization, error) {
	return r.directory.ListOrganizations(ctx)
}
func (r *ScopedRepository) GetOrganization(ctx context.Context, _, id string) (model.Organization, error) {
	return r.directory.GetOrganization(ctx, id)
}
func (r *ScopedRepository) CreateOrganization(ctx context.Context, _ string, input CreateOrganizationInput) (model.Organization, error) {
	return r.directory.CreateOrganization(ctx, input)
}
func (r *ScopedRepository) UpdateOrganization(ctx context.Context, _, id string, input UpdateOrganizationInput) (model.Organization, error) {
	return r.directory.UpdateOrganization(ctx, id, input)
}
func (r *ScopedRepository) ArchiveOrganization(ctx context.Context, _, id string) (model.Organization, error) {
	return r.directory.ArchiveOrganization(ctx, id)
}
func (r *ScopedRepository) CreateVenue(ctx context.Context, _ string, input CreateVenueInput) (model.Venue, error) {
	return r.directory.CreateVenue(ctx, input)
}
func (r *ScopedRepository) UpdateVenue(ctx context.Context, _, id string, input UpdateVenueInput) (model.Venue, error) {
	return r.directory.UpdateVenue(ctx, id, input)
}
func (r *ScopedRepository) ArchiveVenue(ctx context.Context, _, id string) (model.Venue, error) {
	return r.directory.ArchiveVenue(ctx, id)
}

func (r *ScopedRepository) ListPublicVenues(ctx context.Context) ([]model.Venue, error) {
	venues, err := r.directory.ListPublicVenues(ctx)
	if err != nil {
		return nil, err
	}
	return r.withCourts(ctx, venues, true)
}
func (r *ScopedRepository) GetPublicVenue(ctx context.Context, slug string) (model.Venue, error) {
	venue, err := r.directory.GetPublicVenue(ctx, slug)
	if err != nil {
		return model.Venue{}, err
	}
	venue.Courts, err = r.store.ListCourts(ctx, venue.ID, venue.OrganizationID, true)
	return venue, err
}
func (r *ScopedRepository) ListPublicVenueCourts(ctx context.Context, id string) ([]model.Court, error) {
	venue, err := r.directory.GetPublicVenueByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.store.ListCourts(ctx, venue.ID, venue.OrganizationID, true)
}
func (r *ScopedRepository) ListTenantVenues(ctx context.Context, _ string) ([]model.Venue, error) {
	venues, err := r.directory.ListVenues(ctx)
	if err != nil {
		return nil, err
	}
	return r.withCourts(ctx, venues, false)
}
func (r *ScopedRepository) GetTenantVenue(ctx context.Context, _, id string) (model.Venue, error) {
	venue, err := r.directory.AuthorizeVenue(ctx, id, false)
	if errors.Is(err, ErrForbidden) {
		return model.Venue{}, ErrNotFound
	}
	if err != nil {
		return model.Venue{}, err
	}
	venue.Courts, err = r.store.ListCourts(ctx, venue.ID, venue.OrganizationID, false)
	return venue, err
}
func (r *ScopedRepository) ListTenantCourts(ctx context.Context, _, venueID string) ([]model.Court, error) {
	venue, err := r.directory.AuthorizeVenue(ctx, venueID, false)
	if err != nil {
		return nil, err
	}
	return r.store.ListCourts(ctx, venue.ID, venue.OrganizationID, false)
}
func (r *ScopedRepository) GetTenantCourt(ctx context.Context, _, courtID string) (model.Court, error) {
	court, err := r.authorizeCourt(ctx, courtID, false)
	if errors.Is(err, ErrForbidden) {
		return model.Court{}, ErrNotFound
	}
	return court, err
}
func (r *ScopedRepository) CreateCourt(ctx context.Context, _, venueID, name string, interval int32) (model.Court, error) {
	venue, err := r.directory.AuthorizeVenue(ctx, venueID, true)
	if err != nil {
		return model.Court{}, err
	}
	return r.store.CreateCourt(ctx, venue.OrganizationID, venue.ID, name, interval)
}
func (r *ScopedRepository) UpdateCourt(ctx context.Context, _, courtID, name string, interval int32, status string) (model.Court, error) {
	court, err := r.authorizeCourt(ctx, courtID, true)
	if err != nil {
		return model.Court{}, err
	}
	return r.store.UpdateCourt(ctx, court, name, interval, status)
}
func (r *ScopedRepository) ArchiveCourt(ctx context.Context, _, courtID string) (model.Court, error) {
	court, err := r.authorizeCourt(ctx, courtID, true)
	if err != nil {
		return model.Court{}, err
	}
	return r.store.ArchiveCourt(ctx, court)
}
func (r *ScopedRepository) authorizeCourt(ctx context.Context, courtID string, owner bool) (model.Court, error) {
	court, err := r.store.GetCourt(ctx, courtID)
	if err != nil {
		return model.Court{}, err
	}
	venue, err := r.directory.AuthorizeVenue(ctx, court.VenueID, owner)
	if err != nil {
		return model.Court{}, err
	}
	if venue.OrganizationID != court.OrganizationID {
		return model.Court{}, ErrForbidden
	}
	return court, nil
}
func (r *ScopedRepository) withCourts(ctx context.Context, venues []model.Venue, activeOnly bool) ([]model.Venue, error) {
	for i := range venues {
		courts, err := r.store.ListCourts(ctx, venues[i].ID, venues[i].OrganizationID, activeOnly)
		if err != nil {
			return nil, err
		}
		venues[i].Courts = courts
	}
	return venues, nil
}

func (r *ScopedRepository) RequireVenueOwner(ctx context.Context, _, venueID string) error {
	_, err := r.directory.AuthorizeVenue(ctx, venueID, true)
	return err
}
func (r *ScopedRepository) SaveVenueImage(ctx context.Context, _ string, image model.VenueImage) (model.VenueImage, error) {
	if _, err := r.directory.AuthorizeVenue(ctx, image.VenueID, true); err != nil {
		return model.VenueImage{}, err
	}
	return r.store.SaveImage(ctx, image)
}
func (r *ScopedRepository) ListPublicVenueImages(ctx context.Context, venueID string) ([]model.VenueImage, error) {
	if _, err := r.directory.GetPublicVenueByID(ctx, venueID); err != nil {
		return nil, err
	}
	return r.store.ListImages(ctx, venueID)
}
func (r *ScopedRepository) ListTenantVenueImages(ctx context.Context, _, venueID string) ([]model.VenueImage, error) {
	if _, err := r.directory.AuthorizeVenue(ctx, venueID, false); err != nil {
		return nil, err
	}
	return r.store.ListImages(ctx, venueID)
}
func (r *ScopedRepository) GetPublicVenueImage(ctx context.Context, imageID string) (model.VenueImage, error) {
	image, err := r.store.GetImage(ctx, imageID)
	if err != nil {
		return model.VenueImage{}, err
	}
	if _, err := r.directory.GetPublicVenueByID(ctx, image.VenueID); err != nil {
		return model.VenueImage{}, err
	}
	return image, nil
}
func (r *ScopedRepository) GetTenantVenueImage(ctx context.Context, _, imageID string) (model.VenueImage, error) {
	image, err := r.store.GetImage(ctx, imageID)
	if err != nil {
		return model.VenueImage{}, err
	}
	_, err = r.directory.AuthorizeVenue(ctx, image.VenueID, false)
	if errors.Is(err, ErrForbidden) {
		return model.VenueImage{}, ErrNotFound
	}
	if err != nil {
		return model.VenueImage{}, err
	}
	return image, nil
}
