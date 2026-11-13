package staffpostgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

func (r *Repository) ListOrganizationStaff(ctx context.Context, actorID, organizationID string) ([]model.OrganizationStaffDetails, error) {
	actor, organization, err := postgresHelper.ParseTenantScope(actorID, organizationID)
	if err != nil {
		return nil, err
	}

	if _, err := r.queries.FindManagedOrganization(ctx, db.FindManagedOrganizationParams{
		OrganizationID: organization,
		ActorID:        actor,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, identityException.ErrTenantForbidden
		}

		return nil, err
	}

	rows, err := r.queries.ListOrganizationStaff(ctx, organization)
	if err != nil {
		return nil, err
	}

	items := make([]model.OrganizationStaffDetails, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.OrganizationStaffDetails{
			OrganizationStaff: model.OrganizationStaff{
				OrganizationID: row.OrganizationID.String(),
				UserID:         row.UserID.String(),
				Role:           row.Role,
				Status:         row.Status,
			},
			Email:    row.Email,
			FullName: row.FullName,
			VenueIDs: row.VenueIds,
		})
	}

	return items, nil
}
