package authpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	postgresTransaction "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

const usersEmailUniqueConstraint = "users_email_unique_idx"

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

func (r *AuthRepository) CreateUserWithSession(
	ctx context.Context,
	input auth.CreateUserParams,
	session auth.Session,
) (model.User, error) {
	// Parse identifiers before opening the transaction.
	userID, err := uuid.Parse(input.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("parse user ID: %w", err)
	}
	sessionID, err := uuid.Parse(session.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("parse session ID: %w", err)
	}

	var createdUser model.User

	// Persist the user and session atomically.
	err = postgresTransaction.WithinTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.queries.WithTx(tx)

		// Create the user record and map duplicate email errors.
		created, err := queries.CreateUser(ctx, db.CreateUserParams{
			ID:           userID,
			Email:        input.Email,
			PasswordHash: input.PasswordHash,
			FullName:     input.FullName,
			Phone:        input.Phone,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				pgErr.Code == pgerrcode.UniqueViolation &&
				pgErr.ConstraintName == usersEmailUniqueConstraint {
				return auth.ErrEmailTaken
			}
			return fmt.Errorf("create user: %w", err)
		}
		// Create the refresh session for the new user.
		sessionParams := db.CreateAuthSessionParams{
			ID:               sessionID,
			UserID:           userID,
			RefreshTokenHash: session.TokenHash,
			ExpiresAt:        timestamp(session.ExpiresAt),
		}
		if err := queries.CreateAuthSession(ctx, sessionParams); err != nil {
			return fmt.Errorf("create registration session: %w", err)
		}
		createdUser = mapCreatedUser(created)
		return nil
	})
	if err != nil {
		return model.User{}, err
	}
	return createdUser, nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (model.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, auth.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return mapUserByEmail(row), nil
}

func (r *AuthRepository) FindUserByID(ctx context.Context, rawID string) (model.User, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.User{}, auth.ErrUserNotFound
	}
	row, err := r.queries.FindUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, auth.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("find user by ID: %w", err)
	}
	return mapUserByID(row), nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session auth.Session) error {
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
	params := db.CreateAuthSessionParams{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: session.TokenHash,
		ExpiresAt:        timestamp(session.ExpiresAt),
	}
	if err := r.queries.CreateAuthSession(ctx, params); err != nil {
		return fmt.Errorf("create auth session: %w", err)
	}
	return nil
}

func (r *AuthRepository) RotateSession(ctx context.Context, oldHash []byte, next auth.Session, now time.Time) (string, error) {
	// Parse the replacement session identifier before opening the transaction.
	nextID, err := uuid.Parse(next.ID)
	if err != nil {
		return "", fmt.Errorf("parse next session ID: %w", err)
	}

	var userID uuid.UUID

	// Lock, revoke, and replace the session atomically.
	err = postgresTransaction.WithinTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.queries.WithTx(tx)
		lookup := db.FindActiveSessionForUpdateParams{
			RefreshTokenHash: oldHash,
			Now:              timestamp(now),
		}
		userID, err = queries.FindActiveSessionForUpdate(ctx, lookup)
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.ErrInvalidToken
		}
		if err != nil {
			return fmt.Errorf("lock auth session: %w", err)
		}
		// Revoke the previous session before creating its replacement.
		revokeParams := db.RevokeAuthSessionParams{
			RefreshTokenHash: oldHash,
			RevokedAt:        timestamp(now),
		}
		if _, err := queries.RevokeAuthSession(ctx, revokeParams); err != nil {
			return fmt.Errorf("revoke old session: %w", err)
		}
		// Create the replacement session.
		nextParams := db.CreateAuthSessionParams{
			ID:               nextID,
			UserID:           userID,
			RefreshTokenHash: next.TokenHash,
			ExpiresAt:        timestamp(next.ExpiresAt),
		}
		if err := queries.CreateAuthSession(ctx, nextParams); err != nil {
			return fmt.Errorf("create rotated session: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return userID.String(), nil
}

func (r *AuthRepository) RevokeSession(ctx context.Context, hash []byte, now time.Time) error {
	// Revoke the active session associated with the refresh token hash.
	params := db.RevokeAuthSessionParams{
		RefreshTokenHash: hash,
		RevokedAt:        timestamp(now),
	}
	affected, err := r.queries.RevokeAuthSession(ctx, params)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if affected == 0 {
		return auth.ErrInvalidToken
	}
	return nil
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func mapCreatedUser(row db.CreateUserRow) model.User {
	return model.User{
		ID:           row.ID.String(),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		FullName:     row.FullName,
		Phone:        row.Phone,
		Status:       model.Status(row.Status),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

func mapUserByEmail(row db.FindUserByEmailRow) model.User {
	return model.User{
		ID:           row.ID.String(),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		FullName:     row.FullName,
		Phone:        row.Phone,
		Status:       model.Status(row.Status),
		RoleName:     row.RoleName,
		IsAdmin:      row.IsAdmin,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

func mapUserByID(row db.FindUserByIDRow) model.User {
	return model.User{
		ID:           row.ID.String(),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		FullName:     row.FullName,
		Phone:        row.Phone,
		Status:       model.Status(row.Status),
		RoleName:     row.RoleName,
		IsAdmin:      row.IsAdmin,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
