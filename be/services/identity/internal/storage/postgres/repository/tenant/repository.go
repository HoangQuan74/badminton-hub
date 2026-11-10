package tenantpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

type Repository struct{ queries *db.Queries }

func New(pool *pgxpool.Pool) *Repository { return &Repository{queries: db.New(pool)} }

func (r *Repository) ListOrganizations(ctx context.Context, actorID string) ([]model.Organization, error) {
	actor, err := uuid.Parse(actorID)
	if err != nil {
		return nil, tenant.ErrForbidden
	}
	rows, err := r.queries.ListTenantOrganizations(ctx, actor)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	items := make([]model.Organization, 0, len(rows))
	for _, row := range rows {
		items = append(items, organizationModel(row.ID, row.Name, row.Slug, row.Status, row.CreatedAt.Time, row.UpdatedAt.Time))
	}
	return items, nil
}

func (r *Repository) GetOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error) {
	actor, id, err := actorAndID(actorID, organizationID)
	if err != nil {
		return model.Organization{}, err
	}
	row, err := r.queries.GetTenantOrganization(ctx, db.GetTenantOrganizationParams{OrganizationID: id, ActorID: actor})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organization{}, tenant.ErrNotFound
	}
	if err != nil {
		return model.Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return organizationModel(row.ID, row.Name, row.Slug, row.Status, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}

func (r *Repository) CreateOrganization(ctx context.Context, actorID string, input tenant.CreateOrganizationInput) (model.Organization, error) {
	actor, err := uuid.Parse(actorID)
	if err != nil {
		return model.Organization{}, tenant.ErrForbidden
	}
	row, err := r.queries.CreateOrganization(ctx, db.CreateOrganizationParams{ID: uuid.New(), ActorID: actor, Name: input.Name, Slug: input.Slug})
	if err != nil {
		return model.Organization{}, mapWriteError("create organization", err)
	}
	return organizationModel(row.ID, row.Name, row.Slug, row.Status, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}

func (r *Repository) UpdateOrganization(ctx context.Context, actorID, organizationID string, input tenant.UpdateOrganizationInput) (model.Organization, error) {
	actor, id, err := actorAndID(actorID, organizationID)
	if err != nil {
		return model.Organization{}, err
	}
	if err := r.authorizeOrganizationOwner(ctx, actor, id); err != nil {
		return model.Organization{}, err
	}
	row, err := r.queries.UpdateOrganization(ctx, db.UpdateOrganizationParams{OrganizationID: id, ActorID: actor, Name: nullableText(input.Name), Slug: nullableText(input.Slug)})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organization{}, tenant.ErrStateConflict
	}
	if err != nil {
		return model.Organization{}, mapWriteError("update organization", err)
	}
	return organizationModel(row.ID, row.Name, row.Slug, row.Status, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}

func (r *Repository) ArchiveOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error) {
	actor, id, err := actorAndID(actorID, organizationID)
	if err != nil {
		return model.Organization{}, err
	}
	if err := r.authorizeOrganizationOwner(ctx, actor, id); err != nil {
		return model.Organization{}, err
	}
	row, err := r.queries.ArchivePendingOrganization(ctx, db.ArchivePendingOrganizationParams{OrganizationID: id, ActorID: actor})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organization{}, tenant.ErrStateConflict
	}
	if err != nil {
		return model.Organization{}, fmt.Errorf("archive organization: %w", err)
	}
	return organizationModel(row.ID, row.Name, row.Slug, row.Status, row.CreatedAt.Time, row.UpdatedAt.Time), nil
}

func (r *Repository) authorizeOrganizationOwner(ctx context.Context, actorID, organizationID uuid.UUID) error {
	role, err := r.queries.FindOrganizationMembership(ctx, db.FindOrganizationMembershipParams{ActorID: actorID, OrganizationID: organizationID})
	if errors.Is(err, pgx.ErrNoRows) {
		return tenant.ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("check organization membership: %w", err)
	}
	if role != "tenant_owner" {
		return tenant.ErrForbidden
	}
	return nil
}

func actorAndID(actorID, rawID string) (uuid.UUID, uuid.UUID, error) {
	actor, err := uuid.Parse(actorID)
	if err != nil {
		return uuid.Nil, uuid.Nil, tenant.ErrForbidden
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.Nil, uuid.Nil, tenant.ErrInvalidInput
	}
	return actor, id, nil
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func mapWriteError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "identity_organizations_slug_unique_idx" {
		return tenant.ErrSlugConflict
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "identity_venues_slug_unique_idx" {
		return tenant.ErrVenueSlugConflict
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return tenant.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func organizationModel(id uuid.UUID, name, slug, status string, createdAt, updatedAt time.Time) model.Organization {
	return model.Organization{ID: id.String(), Name: name, Slug: slug, Status: status, CreatedAt: createdAt, UpdatedAt: updatedAt}
}
