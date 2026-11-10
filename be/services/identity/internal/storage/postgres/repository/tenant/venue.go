package tenantpostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

func (r *Repository) ListPublicVenues(ctx context.Context) ([]model.Venue, error) {
	rows, err := r.queries.ListIdentityPublicVenues(ctx)
	if err != nil {
		return nil, fmt.Errorf("list public venues: %w", err)
	}
	items := make([]model.Venue, 0, len(rows))
	for _, row := range rows {
		items = append(items, venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status))
	}
	return items, nil
}
func (r *Repository) GetPublicVenue(ctx context.Context, slug string) (model.Venue, error) {
	row, err := r.queries.GetIdentityPublicVenueBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Venue{}, tenant.ErrNotFound
	}
	if err != nil {
		return model.Venue{}, fmt.Errorf("get public venue: %w", err)
	}
	return venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status), nil
}
func (r *Repository) GetPublicVenueByID(ctx context.Context, venueID string) (model.Venue, error) {
	id, err := uuid.Parse(venueID)
	if err != nil {
		return model.Venue{}, tenant.ErrInvalidInput
	}
	row, err := r.queries.GetIdentityPublicVenueByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Venue{}, tenant.ErrNotFound
	}
	if err != nil {
		return model.Venue{}, fmt.Errorf("get public venue by ID: %w", err)
	}
	return venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status), nil
}
func (r *Repository) ListVenues(ctx context.Context, actorID string) ([]model.Venue, error) {
	actor, err := uuid.Parse(actorID)
	if err != nil {
		return nil, tenant.ErrForbidden
	}
	rows, err := r.queries.ListIdentityTenantVenues(ctx, actor)
	if err != nil {
		return nil, fmt.Errorf("list tenant venues: %w", err)
	}
	items := make([]model.Venue, 0, len(rows))
	for _, row := range rows {
		items = append(items, venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status))
	}
	return items, nil
}
func (r *Repository) AuthorizeVenue(ctx context.Context, actorID, venueID string, requireOwner bool) (model.Venue, error) {
	actor, id, err := actorAndID(actorID, venueID)
	if err != nil {
		return model.Venue{}, err
	}
	row, err := r.queries.GetIdentityAuthorizedVenue(ctx, db.GetIdentityAuthorizedVenueParams{VenueID: id, ActorID: actor, RequireOwner: requireOwner})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Venue{}, tenant.ErrForbidden
	}
	if err != nil {
		return model.Venue{}, fmt.Errorf("authorize venue: %w", err)
	}
	return venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status), nil
}
func (r *Repository) CreateVenue(ctx context.Context, actorID string, input tenant.CreateVenueInput) (model.Venue, error) {
	actor, organization, err := actorAndID(actorID, input.OrganizationID)
	if err != nil {
		return model.Venue{}, err
	}
	row, err := r.queries.CreateIdentityVenue(ctx, db.CreateIdentityVenueParams{ID: uuid.New(), ActorID: actor, OrganizationID: organization, Name: input.Name, Slug: input.Slug, Address: input.Address, Timezone: input.Timezone})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Venue{}, tenant.ErrForbidden
	}
	if err != nil {
		return model.Venue{}, mapWriteError("create venue", err)
	}
	return venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status), nil
}
func (r *Repository) UpdateVenue(ctx context.Context, actorID, venueID string, input tenant.UpdateVenueInput) (model.Venue, error) {
	actor, id, err := actorAndID(actorID, venueID)
	if err != nil {
		return model.Venue{}, err
	}
	if _, err := r.AuthorizeVenue(ctx, actorID, venueID, true); err != nil {
		return model.Venue{}, err
	}
	row, err := r.queries.UpdateIdentityVenue(ctx, db.UpdateIdentityVenueParams{VenueID: id, ActorID: actor, Name: nullableText(input.Name), Slug: nullableText(input.Slug), Address: nullableText(input.Address), Timezone: nullableText(input.Timezone)})
	if err != nil {
		return model.Venue{}, mapWriteError("update venue", err)
	}
	return venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status), nil
}
func (r *Repository) ArchiveVenue(ctx context.Context, actorID, venueID string) (model.Venue, error) {
	actor, id, err := actorAndID(actorID, venueID)
	if err != nil {
		return model.Venue{}, err
	}
	if _, err := r.AuthorizeVenue(ctx, actorID, venueID, true); err != nil {
		return model.Venue{}, err
	}
	row, err := r.queries.ArchiveIdentityVenue(ctx, db.ArchiveIdentityVenueParams{VenueID: id, ActorID: actor})
	if err != nil {
		return model.Venue{}, mapWriteError("archive venue", err)
	}
	return venueModel(row.ID, row.OrganizationID, row.Name, row.Slug, row.Address, row.Timezone, row.Status), nil
}
func venueModel(id, organizationID uuid.UUID, name, slug, address, timezone, status string) model.Venue {
	return model.Venue{ID: id.String(), OrganizationID: organizationID.String(), Name: name, Slug: slug, Address: address, Timezone: timezone, Status: status}
}
