package tenantpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *Repository) ListOrganizations(ctx context.Context, actorID string) ([]model.Organization, error) {
	actor, err := uuid.Parse(actorID)
	if err != nil {
		return nil, identityException.ErrTenantForbidden
	}

	rows, err := r.queries.ListTenantOrganizations(ctx, actor)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}

	items := make([]model.Organization, 0, len(rows))
	for _, row := range rows {
		items = append(items, organizationModel(
			row.ID,
			row.Name,
			row.Slug,
			row.Status,
			row.CreatedAt.Time,
			row.UpdatedAt.Time,
		))
	}

	return items, nil
}

func (r *Repository) GetOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error) {
	actor, id, err := postgresHelper.ParseTenantScope(actorID, organizationID)
	if err != nil {
		return model.Organization{}, err
	}

	row, err := r.queries.GetTenantOrganization(ctx, db.GetTenantOrganizationParams{
		OrganizationID: id,
		ActorID:        actor,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organization{}, identityException.ErrTenantNotFound
	}

	if err != nil {
		return model.Organization{}, fmt.Errorf("get organization: %w", err)
	}

	return organizationModel(
		row.ID,
		row.Name,
		row.Slug,
		row.Status,
		row.CreatedAt.Time,
		row.UpdatedAt.Time,
	), nil
}

func (r *Repository) UpdateOrganization(ctx context.Context, actorID, organizationID string, input tenant.UpdateOrganizationInput) (model.Organization, error) {
	actor, id, err := postgresHelper.ParseTenantScope(actorID, organizationID)
	if err != nil {
		return model.Organization{}, err
	}

	row, err := r.queries.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		OrganizationID: id,
		ActorID:        actor,
		Name:           postgresHelper.MapOptionalText(input.Name),
		Slug:           postgresHelper.MapOptionalText(input.Slug),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organization{}, identityException.ErrTenantStateConflict
	}

	if err != nil {
		return model.Organization{}, identityException.TenantWriteError("update organization", err)
	}

	return organizationModel(
		row.ID,
		row.Name,
		row.Slug,
		row.Status,
		row.CreatedAt.Time,
		row.UpdatedAt.Time,
	), nil
}

func (r *Repository) ArchiveOrganization(ctx context.Context, actorID, organizationID string) (model.Organization, error) {
	actor, id, err := postgresHelper.ParseTenantScope(actorID, organizationID)
	if err != nil {
		return model.Organization{}, err
	}

	row, err := r.queries.ArchivePendingOrganization(ctx, db.ArchivePendingOrganizationParams{
		OrganizationID: id,
		ActorID:        actor,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Organization{}, identityException.ErrTenantStateConflict
	}

	if err != nil {
		return model.Organization{}, fmt.Errorf("archive organization: %w", err)
	}

	return organizationModel(
		row.ID,
		row.Name,
		row.Slug,
		row.Status,
		row.CreatedAt.Time,
		row.UpdatedAt.Time,
	), nil
}

func (r *Repository) GetOrganizationRole(ctx context.Context, actorID, organizationID string) (string, error) {
	actor, organization, err := postgresHelper.ParseTenantScope(actorID, organizationID)
	if err != nil {
		return "", err
	}

	role, err := r.queries.FindOrganizationMembership(ctx, db.FindOrganizationMembershipParams{
		ActorID:        actor,
		OrganizationID: organization,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", identityException.ErrTenantForbidden
	}

	return role, err
}

func organizationModel(id uuid.UUID, name, slug, status string, createdAt, updatedAt time.Time) model.Organization {
	return model.Organization{
		ID:        id.String(),
		Name:      name,
		Slug:      slug,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
