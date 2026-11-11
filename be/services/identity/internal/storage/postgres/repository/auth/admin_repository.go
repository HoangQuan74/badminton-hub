package authpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	postgresTransaction "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

func (r *AuthRepository) FindAdminByEmail(ctx context.Context, email string) (model.AdminCredentials, error) {
	row, err := r.queries.FindAdminByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AdminCredentials{}, auth.ErrInvalidCredentials
	}
	if err != nil {
		return model.AdminCredentials{}, fmt.Errorf("find admin: %w", err)
	}
	return model.AdminCredentials{
		Staff: model.Staff{
			ID: row.ID.String(), Email: row.Email, FullName: row.FullName,
			Status: row.Status, IsAdmin: row.IsAdmin, RoleID: row.RoleID, RoleName: row.RoleName,
		},
		PasswordHash: row.PasswordHash,
	}, nil
}

func (r *AuthRepository) RotateAdminSession(ctx context.Context, hash []byte, next auth.Session) (string, error) {
	nextID, err := uuid.Parse(next.ID)
	if err != nil {
		return "", err
	}
	var userID uuid.UUID
	// Lock the active admin session and rotate it atomically.
	err = postgresTransaction.WithinTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.queries.WithTx(tx)
		var err error
		userID, err = queries.FindAdminSessionForUpdate(ctx, hash)
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.ErrInvalidToken
		}
		if err != nil {
			return fmt.Errorf("lock admin session: %w", err)
		}
		revoke := db.RevokeAuthSessionParams{
			RefreshTokenHash: hash,
			RevokedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}
		_, err = queries.RevokeAuthSession(ctx, revoke)
		if err != nil {
			return fmt.Errorf("revoke admin session: %w", err)
		}
		replacement := db.CreateAuthSessionParams{
			ID: nextID, UserID: userID, RefreshTokenHash: next.TokenHash,
			ExpiresAt: pgtype.Timestamptz{Time: next.ExpiresAt, Valid: true},
		}
		return queries.CreateAuthSession(ctx, replacement)
	})
	if err != nil {
		return "", err
	}
	return userID.String(), nil
}

func (r *AuthRepository) RevokeAdminSession(ctx context.Context, hash []byte) error {
	params := db.RevokeAuthSessionParams{
		RefreshTokenHash: hash,
		RevokedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	_, err := r.queries.RevokeAuthSession(ctx, params)
	return err
}

func (r *AuthRepository) PasswordHash(ctx context.Context, rawID string) (string, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return "", auth.ErrUserNotFound
	}
	hash, err := r.queries.GetAdminPasswordHash(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", auth.ErrUserNotFound
	}
	return hash, err
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, rawID, hash string) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return auth.ErrUserNotFound
	}
	return r.queries.UpdateAdminPassword(ctx, db.UpdateAdminPasswordParams{ID: id, PasswordHash: hash})
}
