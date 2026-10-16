package postgres

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
	"github.com/kainguyen/badminton-hub/be/services/identity/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/domain/user"
	db "github.com/kainguyen/badminton-hub/be/services/identity/storage/sqlc"
)

type AuthRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool, queries: db.New(pool)}
}

func (r *AuthRepository) CreateUserWithSession(ctx context.Context, input auth.CreateUserParams, session auth.Session) (user.User, error) {
	userID, sessionID, err := parseIDs(input.ID, session.ID)
	if err != nil {
		return user.User{}, err
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return user.User{}, fmt.Errorf("begin user registration: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	created, err := queries.CreateUser(ctx, db.CreateUserParams{
		ID: userID, Email: input.Email, PasswordHash: input.PasswordHash, FullName: input.FullName, Phone: input.Phone,
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return user.User{}, auth.ErrEmailTaken
	}
	if err != nil {
		return user.User{}, fmt.Errorf("create user: %w", err)
	}
	if err := queries.CreateAuthSession(ctx, db.CreateAuthSessionParams{
		ID: sessionID, UserID: userID, RefreshTokenHash: session.TokenHash, ExpiresAt: timestamp(session.ExpiresAt),
	}); err != nil {
		return user.User{}, fmt.Errorf("create registration session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return user.User{}, fmt.Errorf("commit user registration: %w", err)
	}
	return mapCreatedUser(created), nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (user.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, auth.ErrUserNotFound
	}
	if err != nil {
		return user.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return mapUserByEmail(row), nil
}

func (r *AuthRepository) FindUserByID(ctx context.Context, rawID string) (user.User, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return user.User{}, auth.ErrUserNotFound
	}
	row, err := r.queries.FindUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, auth.ErrUserNotFound
	}
	if err != nil {
		return user.User{}, fmt.Errorf("find user by ID: %w", err)
	}
	return mapUserByID(row), nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session auth.Session) error {
	sessionID, userID, err := parseIDs(session.ID, session.UserID)
	if err != nil {
		return err
	}
	if err := r.queries.CreateAuthSession(ctx, db.CreateAuthSessionParams{
		ID: sessionID, UserID: userID, RefreshTokenHash: session.TokenHash, ExpiresAt: timestamp(session.ExpiresAt),
	}); err != nil {
		return fmt.Errorf("create auth session: %w", err)
	}
	return nil
}

func (r *AuthRepository) RotateSession(ctx context.Context, oldHash []byte, next auth.Session, now time.Time) (string, error) {
	nextID, err := uuid.Parse(next.ID)
	if err != nil {
		return "", fmt.Errorf("parse next session ID: %w", err)
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("begin session rotation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	userID, err := queries.FindActiveSessionForUpdate(ctx, db.FindActiveSessionForUpdateParams{RefreshTokenHash: oldHash, Now: timestamp(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", auth.ErrInvalidToken
	}
	if err != nil {
		return "", fmt.Errorf("lock auth session: %w", err)
	}
	if _, err := queries.RevokeAuthSession(ctx, db.RevokeAuthSessionParams{RefreshTokenHash: oldHash, RevokedAt: timestamp(now)}); err != nil {
		return "", fmt.Errorf("revoke old session: %w", err)
	}
	if err := queries.CreateAuthSession(ctx, db.CreateAuthSessionParams{
		ID: nextID, UserID: userID, RefreshTokenHash: next.TokenHash, ExpiresAt: timestamp(next.ExpiresAt),
	}); err != nil {
		return "", fmt.Errorf("create rotated session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit session rotation: %w", err)
	}
	return userID.String(), nil
}

func (r *AuthRepository) RevokeSession(ctx context.Context, hash []byte, now time.Time) error {
	affected, err := r.queries.RevokeAuthSession(ctx, db.RevokeAuthSessionParams{RefreshTokenHash: hash, RevokedAt: timestamp(now)})
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if affected == 0 {
		return auth.ErrInvalidToken
	}
	return nil
}

func parseIDs(first, second string) (uuid.UUID, uuid.UUID, error) {
	firstID, err := uuid.Parse(first)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("parse UUID: %w", err)
	}
	secondID, err := uuid.Parse(second)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("parse UUID: %w", err)
	}
	return firstID, secondID, nil
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func mapCreatedUser(row db.CreateUserRow) user.User {
	return user.User{ID: row.ID.String(), Email: row.Email, PasswordHash: row.PasswordHash, FullName: row.FullName, Phone: row.Phone, Status: user.Status(row.Status), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func mapUserByEmail(row db.FindUserByEmailRow) user.User {
	return user.User{ID: row.ID.String(), Email: row.Email, PasswordHash: row.PasswordHash, FullName: row.FullName, Phone: row.Phone, Status: user.Status(row.Status), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}

func mapUserByID(row db.FindUserByIDRow) user.User {
	return user.User{ID: row.ID.String(), Email: row.Email, PasswordHash: row.PasswordHash, FullName: row.FullName, Phone: row.Phone, Status: user.Status(row.Status), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}
