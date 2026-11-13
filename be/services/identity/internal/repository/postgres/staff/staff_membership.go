package staffpostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

func (r *Repository) SetOrganizationStaff(ctx context.Context, actorID string, input staff.SetOrganizationStaffInput) (model.OrganizationStaff, error) {
	actor, organization, err := postgresHelper.ParseTenantScope(actorID, input.OrganizationID)
	if err != nil {
		return model.OrganizationStaff{}, err
	}

	user, err := uuid.Parse(input.UserID)
	if err != nil {
		return model.OrganizationStaff{}, identityException.ErrTenantInvalidInput
	}

	row, err := r.queries.SetIdentityOrganizationStaff(ctx, db.SetIdentityOrganizationStaffParams{
		ActorID:        actor,
		OrganizationID: organization,
		UserID:         user,
		Role:           input.Role,
		Status:         input.Status,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.OrganizationStaff{}, identityException.ErrTenantForbidden
	}

	if err != nil {
		return model.OrganizationStaff{}, fmt.Errorf("set tenant staff: %w", err)
	}

	return model.OrganizationStaff{
		OrganizationID: row.OrganizationID.String(),
		UserID:         row.UserID.String(),
		Role:           row.Role,
		Status:         row.Status,
	}, nil
}

func (r *Repository) SetVenueStaff(ctx context.Context, actorID string, input staff.SetVenueStaffInput) (model.VenueStaff, error) {
	actor, venue, err := postgresHelper.ParseTenantScope(actorID, input.VenueID)
	if err != nil {
		return model.VenueStaff{}, err
	}

	user, err := uuid.Parse(input.UserID)
	if err != nil {
		return model.VenueStaff{}, identityException.ErrTenantInvalidInput
	}

	row, err := r.queries.SetIdentityVenueStaff(ctx, db.SetIdentityVenueStaffParams{
		ActorID: actor,
		VenueID: venue,
		UserID:  user,
		Status:  input.Status,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.VenueStaff{}, identityException.ErrTenantForbidden
	}

	if err != nil {
		return model.VenueStaff{}, fmt.Errorf("set venue staff: %w", err)
	}

	return model.VenueStaff{
		OrganizationID: row.OrganizationID.String(),
		VenueID:        row.VenueID.String(),
		UserID:         row.UserID.String(),
		Status:         row.Status,
	}, nil
}
