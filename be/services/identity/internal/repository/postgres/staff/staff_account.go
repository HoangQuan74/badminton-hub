package staffpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	postgresTransaction "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

func (r *Repository) FindStaffCredentialsByEmail(ctx context.Context, email string) (model.StaffCredentials, error) {
	row, err := r.queries.FindStaffCredentialsByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.StaffCredentials{}, identityException.ErrAuthInvalidCredentials
	}

	if err != nil {
		return model.StaffCredentials{}, fmt.Errorf("find admin: %w", err)
	}

	return model.StaffCredentials{
		Staff: model.Staff{
			ID:       row.ID.String(),
			Email:    row.Email,
			FullName: row.FullName,
			Status:   row.Status,
			IsAdmin:  row.IsAdmin,
			RoleID:   row.RoleID,
			RoleName: row.RoleName,
		},
		PasswordHash: row.PasswordHash,
	}, nil
}

func (r *Repository) RotateStaffSession(ctx context.Context, hash []byte, next model.Session) (string, error) {
	nextID, err := uuid.Parse(next.ID)
	if err != nil {
		return "", err
	}

	var userID uuid.UUID
	// Lock the active admin session and rotate it atomically.
	err = postgresTransaction.WithinTransaction(
		ctx,
		r.pool,
		func(tx pgx.Tx) error {
			queries := r.queries.WithTx(tx)
			var err error
			userID, err = queries.FindStaffSessionForUpdate(ctx, hash)
			if errors.Is(err, pgx.ErrNoRows) {
				return identityException.ErrAuthInvalidToken
			}

			if err != nil {
				return fmt.Errorf("lock admin session: %w", err)
			}

			revoke := db.RevokeAuthSessionParams{
				RefreshTokenHash: hash,
				RevokedAt:        postgresHelper.MapTimestamp(time.Now().UTC()),
			}
			_, err = queries.RevokeAuthSession(ctx, revoke)
			if err != nil {
				return fmt.Errorf("revoke admin session: %w", err)
			}

			replacement := postgresHelper.MapSessionParams(
				nextID,
				userID,
				next,
			)
			return queries.CreateAuthSession(ctx, replacement)
		},
	)
	if err != nil {
		return "", err
	}

	return userID.String(), nil
}

func (r *Repository) RevokeStaffSession(ctx context.Context, hash []byte) error {
	params := db.RevokeAuthSessionParams{
		RefreshTokenHash: hash,
		RevokedAt:        postgresHelper.MapTimestamp(time.Now().UTC()),
	}
	_, err := r.queries.RevokeAuthSession(ctx, params)
	return err
}

func (r *Repository) GetStaffPasswordHash(ctx context.Context, rawID string) (string, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return "", identityException.ErrAuthUserNotFound
	}

	hash, err := r.queries.GetStaffPasswordHash(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", identityException.ErrAuthUserNotFound
	}

	return hash, err
}

func (r *Repository) UpdateStaffPassword(ctx context.Context, rawID, hash string) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return identityException.ErrAuthUserNotFound
	}

	return r.queries.UpdateStaffPassword(ctx, db.UpdateStaffPasswordParams{
		ID:           id,
		PasswordHash: hash,
	})
}

func (r *Repository) CreateStaffSession(ctx context.Context, session model.Session) error {
	id, err := uuid.Parse(session.ID)
	if err != nil {
		return err
	}

	user, err := uuid.Parse(session.UserID)
	if err != nil {
		return err
	}

	return r.queries.CreateAuthSession(ctx, postgresHelper.MapSessionParams(
		id,
		user,
		session,
	))
}
