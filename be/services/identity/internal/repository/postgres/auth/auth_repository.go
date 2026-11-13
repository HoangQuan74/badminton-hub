package authpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	postgresTransaction "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

type AuthRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *AuthRepository) FindCredentialsByEmail(ctx context.Context, email string) (model.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, identityException.ErrAuthUserNotFound
	}

	if err != nil {
		return model.User{}, fmt.Errorf("find user by email: %w", err)
	}

	return postgresHelper.MapPrincipal(db.FindUserByIDRow(row)), nil
}

func (r *AuthRepository) GetPrincipal(ctx context.Context, rawID string) (model.User, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.User{}, identityException.ErrAuthUserNotFound
	}

	row, err := r.queries.FindUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, identityException.ErrAuthUserNotFound
	}

	if err != nil {
		return model.User{}, fmt.Errorf("find user by ID: %w", err)
	}

	return postgresHelper.MapPrincipal(row), nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session model.Session) error {
	// Parse session identifiers before persistence.
	sessionID, err := uuid.Parse(session.ID)
	if err != nil {
		return fmt.Errorf("parse session ID: %w", err)
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		return fmt.Errorf("parse user ID: %w", err)
	}

	// Persist the refresh session.
	params := postgresHelper.MapSessionParams(
		sessionID,
		userID,
		session,
	)
	if err := r.queries.CreateAuthSession(ctx, params); err != nil {
		return fmt.Errorf("create auth session: %w", err)
	}

	return nil
}

type sessionTransaction struct{ *AuthRepository }

func (r *AuthRepository) WithinSessionTransaction(ctx context.Context, execute func(auth.AuthSessionTransaction) error) error {
	return postgresTransaction.WithinTransaction(
		ctx,
		r.pool,
		func(tx pgx.Tx) error {
			repository := &AuthRepository{queries: r.queries.WithTx(tx)}
			return execute(&sessionTransaction{AuthRepository: repository})
		},
	)
}

func (t *sessionTransaction) LockSessionOwner(ctx context.Context, hash []byte, now time.Time) (string, error) {
	id, err := t.queries.FindActiveSessionForUpdate(ctx, db.FindActiveSessionForUpdateParams{
		RefreshTokenHash: hash,
		Now:              postgresHelper.MapTimestamp(now),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", identityException.ErrAuthInvalidToken
	}

	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func (r *AuthRepository) RevokeSession(ctx context.Context, hash []byte, now time.Time) error {
	// Revoke the active session associated with the refresh token hash.
	params := db.RevokeAuthSessionParams{
		RefreshTokenHash: hash,
		RevokedAt:        postgresHelper.MapTimestamp(now),
	}
	affected, err := r.queries.RevokeAuthSession(ctx, params)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	if affected == 0 {
		return identityException.ErrAuthInvalidToken
	}

	return nil
}
