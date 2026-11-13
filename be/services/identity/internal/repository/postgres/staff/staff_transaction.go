package staffpostgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

type membershipTransaction struct{ queries *db.Queries }

func (r *Repository) WithinMembershipTransaction(ctx context.Context, execute func(staff.StaffMembershipTransaction) error) error {
	return transaction.WithinTransaction(
		ctx,
		r.pool,
		func(tx pgx.Tx) error {
			return execute(&membershipTransaction{queries: r.queries.WithTx(tx)})
		},
	)
}

func (t *membershipTransaction) GetManagedOrganizationStatus(ctx context.Context, actorID, organizationID string) (string, error) {
	actor, organization, err := postgresHelper.ParseTenantScope(actorID, organizationID)
	if err != nil {
		return "", err
	}

	status, err := t.queries.FindManagedOrganization(ctx, db.FindManagedOrganizationParams{
		OrganizationID: organization,
		ActorID:        actor,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", identityException.ErrTenantForbidden
	}

	return status, err
}

func (t *membershipTransaction) LockCandidate(ctx context.Context, email string) (model.User, error) {
	row, err := t.queries.LockStaffCandidate(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, identityException.ErrTenantNotFound
	}

	if err != nil {
		return model.User{}, err
	}

	return model.User{
		ID:       row.ID.String(),
		Status:   model.AccountStatus(row.Status),
		IsAdmin:  row.IsAdmin,
		RoleName: row.RoleName,
	}, nil
}

func (t *membershipTransaction) AssignRole(ctx context.Context, userID, role string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return identityException.ErrTenantInvalidInput
	}

	count, err := t.queries.AssignStaffRole(ctx, db.AssignStaffRoleParams{
		Role:   role,
		UserID: id,
	})
	if err != nil {
		return err
	}

	if count != 1 {
		return identityException.ErrTenantStateConflict
	}

	return nil
}

func (t *membershipTransaction) SetOrganizationStaff(ctx context.Context, actorID string, input staff.SetOrganizationStaffInput) (model.OrganizationStaff, error) {
	return (&Repository{queries: t.queries}).SetOrganizationStaff(
		ctx,
		actorID,
		input,
	)
}

func (t *membershipTransaction) DeactivateOrganizationStaff(ctx context.Context, actorID string, input staff.RemoveOrganizationStaffInput) (model.OrganizationStaff, error) {
	actor, organization, err := postgresHelper.ParseTenantScope(actorID, input.OrganizationID)
	if err != nil {
		return model.OrganizationStaff{}, err
	}

	user, err := uuid.Parse(input.UserID)
	if err != nil {
		return model.OrganizationStaff{}, identityException.ErrTenantInvalidInput
	}

	row, err := t.queries.DeactivateOrganizationStaff(ctx, db.DeactivateOrganizationStaffParams{
		OrganizationID: organization,
		UserID:         user,
		ActorID:        actor,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.OrganizationStaff{}, identityException.ErrTenantForbidden
	}

	if err != nil {
		return model.OrganizationStaff{}, err
	}

	return model.OrganizationStaff{
		OrganizationID: row.OrganizationID.String(),
		UserID:         row.UserID.String(),
		Role:           row.Role,
		Status:         row.Status,
	}, nil
}

func (t *membershipTransaction) DeactivateMemberVenues(ctx context.Context, organizationID, userID string) error {
	organization, user, err := postgresHelper.ParseTenantScope(organizationID, userID)
	if err != nil {
		return err
	}

	return t.queries.DeactivateMemberVenues(ctx, db.DeactivateMemberVenuesParams{
		OrganizationID: organization,
		UserID:         user,
	})
}
