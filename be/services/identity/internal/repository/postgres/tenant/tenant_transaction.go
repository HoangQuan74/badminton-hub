package tenantpostgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

type tenantTransaction struct{ queries *db.Queries }

func (r *Repository) WithinCreationTransaction(ctx context.Context, execute func(tenant.TenantCreationTransaction) error) error {
	err := transaction.WithinTransaction(
		ctx,
		r.pool,
		func(tx pgx.Tx) error {
			return execute(&tenantTransaction{queries: r.queries.WithTx(tx)})
		},
	)
	return identityException.TenantPersistenceError(err)
}

func (r *Repository) WithinReviewTransaction(ctx context.Context, execute func(tenant.TenantReviewTransaction) error) error {
	return transaction.WithinTransaction(
		ctx,
		r.pool,
		func(tx pgx.Tx) error {
			return execute(&tenantTransaction{queries: r.queries.WithTx(tx)})
		},
	)
}

func (t *tenantTransaction) LockCreator(ctx context.Context, actorID string) (model.User, error) {
	id, err := uuid.Parse(actorID)
	if err != nil {
		return model.User{}, identityException.ErrTenantUnauthorized
	}

	row, err := t.queries.LockTenantCreator(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, identityException.ErrTenantUnauthorized
	}

	if err != nil {
		return model.User{}, err
	}

	return model.User{
		ID:        row.ID.String(),
		Email:     row.Email,
		FullName:  row.FullName,
		Phone:     row.Phone,
		Status:    model.AccountStatus(row.Status),
		RoleName:  row.RoleName,
		IsAdmin:   row.IsAdmin,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (t *tenantTransaction) FindCreateReplay(ctx context.Context, actorID, key string) (tenant.CreationReplay, bool, error) {
	actor, requestKey, err := postgresHelper.ParseTenantScope(actorID, key)
	if err != nil {
		return tenant.CreationReplay{}, false, err
	}

	row, err := t.queries.FindTenantCreateReplay(ctx, db.FindTenantCreateReplayParams{
		ActorID:        actor,
		IdempotencyKey: requestKey,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return tenant.CreationReplay{}, false, nil
	}

	if err != nil {
		return tenant.CreationReplay{}, false, err
	}

	var output tenant.CreateOrganizationOutput
	if err := json.Unmarshal(row.Response, &output); err != nil {
		return tenant.CreationReplay{}, false, fmt.Errorf("decode tenant replay: %w", err)
	}

	output.ExpiresAt = row.ExpiresAt.Time
	return tenant.CreationReplay{
		Output:      output,
		RequestHash: row.RequestHash,
	}, true, nil
}

func (t *tenantTransaction) DeleteExpiredCreateKey(ctx context.Context, actorID, key string) error {
	actor, requestKey, err := postgresHelper.ParseTenantScope(actorID, key)
	if err != nil {
		return err
	}

	return t.queries.DeleteExpiredTenantCreateKey(ctx, db.DeleteExpiredTenantCreateKeyParams{
		ActorID:        actor,
		IdempotencyKey: requestKey,
	})
}

func (t *tenantTransaction) SetOwnerRole(ctx context.Context, actorID string) error {
	id, err := uuid.Parse(actorID)
	if err != nil {
		return identityException.ErrTenantUnauthorized
	}

	count, err := t.queries.SetTenantOwnerRole(ctx, id)
	if err != nil {
		return err
	}

	if count != 1 {
		return identityException.ErrTenantOwnerRoleMissing
	}

	return nil
}

func (t *tenantTransaction) CreateOrganization(ctx context.Context, input tenant.CreateOrganizationParams) (model.Organization, error) {
	actor, err := uuid.Parse(input.ActorID)
	if err != nil {
		return model.Organization{}, identityException.ErrTenantUnauthorized
	}

	row, err := t.queries.CreateOrganization(ctx, db.CreateOrganizationParams{
		ID:      uuid.New(),
		ActorID: actor,
		Name:    input.Name,
		Slug:    input.Slug,
	})
	if err != nil {
		return model.Organization{}, err
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

func (t *tenantTransaction) SaveRegistration(ctx context.Context, organizationID string, input tenant.RegistrationContactInput) error {
	id, err := uuid.Parse(organizationID)
	if err != nil {
		return identityException.ErrTenantInvalidInput
	}

	return t.queries.SaveTenantRegistration(ctx, db.SaveTenantRegistrationParams{
		OrganizationID: id,
		ContactName:    input.ContactName,
		ContactEmail:   input.ContactEmail,
		ContactPhone:   input.ContactPhone,
		Note:           input.Note,
	})
}

func (t *tenantTransaction) SaveCreateReplay(ctx context.Context, input tenant.CreateOrganizationParams, output tenant.CreateOrganizationOutput) (time.Time, error) {
	actor, key, err := postgresHelper.ParseTenantScope(input.ActorID, input.Key)
	if err != nil {
		return time.Time{}, err
	}

	snapshot, err := json.Marshal(output)
	if err != nil {
		return time.Time{}, err
	}

	expires, err := t.queries.SaveTenantCreateReplay(ctx, db.SaveTenantCreateReplayParams{
		ActorID:        actor,
		IdempotencyKey: key,
		RequestHash:    input.RequestHash,
		Response:       snapshot,
	})
	return expires.Time, err
}

func (t *tenantTransaction) LockRegistration(ctx context.Context, organizationID string) (model.TenantRegistration, error) {
	id, err := uuid.Parse(organizationID)
	if err != nil {
		return model.TenantRegistration{}, identityException.ErrTenantInvalidInput
	}

	if _, err := t.queries.LockTenantRegistration(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.TenantRegistration{}, identityException.ErrTenantNotFound
		}

		return model.TenantRegistration{}, err
	}

	return t.GetRegistration(ctx, organizationID)
}

func (t *tenantTransaction) SaveRegistrationReview(ctx context.Context, actorID string, input tenant.ReviewRegistrationInput) error {
	actor, id, err := postgresHelper.ParseTenantScope(actorID, input.OrganizationID)
	if err != nil {
		return err
	}

	if err := t.queries.ApplyTenantRegistrationDecision(ctx, db.ApplyTenantRegistrationDecisionParams{
		ID:     id,
		Status: input.Decision,
	}); err != nil {
		return err
	}

	return t.queries.RecordTenantRegistrationReview(ctx, db.RecordTenantRegistrationReviewParams{
		OrganizationID: id,
		ReviewerID:     actor,
		Decision:       input.Decision,
		Note:           input.Note,
	})
}

func (t *tenantTransaction) GetRegistration(ctx context.Context, organizationID string) (model.TenantRegistration, error) {
	id, err := uuid.Parse(organizationID)
	if err != nil {
		return model.TenantRegistration{}, identityException.ErrTenantInvalidInput
	}

	row, err := t.queries.GetTenantRegistration(ctx, id)
	if err != nil {
		return model.TenantRegistration{}, err
	}

	return registrationModel(row), nil
}
