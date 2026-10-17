package userpostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	postgresTransaction "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: db.New(pool)}
}

func (r *Repository) List(ctx context.Context, query model.ListQuery) ([]model.User, error) {
	// Load the requested user page from PostgreSQL.
	queryParams := db.ListUsersParams{
		LimitCount:  int32(query.Limit),
		OffsetCount: int32(query.Offset),
		Search:      query.Search,
		CreatedFrom: postgresTransaction.TimestampPointer(query.CreatedFrom),
		CreatedTo:   postgresTransaction.TimestampPointer(query.CreatedTo),
	}
	rows, err := r.queries.ListUsers(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	// Map storage rows to reusable domain models.
	users := make([]model.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, mapListUser(row))
	}
	return users, nil
}

func (r *Repository) FindByID(ctx context.Context, rawID string) (model.User, error) {
	// Parse the user identifier before querying PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.User{}, userDomain.ErrInvalidInput
	}
	row, err := r.queries.FindUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, userDomain.ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("find user: %w", err)
	}
	return mapUser(row), nil
}

func (r *Repository) Create(ctx context.Context, input userDomain.CreateParams) (model.User, error) {
	// Create a managed user record.
	queryParams := db.CreateManagedUserParams{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		FullName:     input.FullName,
		Phone:        input.Phone,
		Status:       input.Status,
	}
	row, err := r.queries.CreateManagedUser(ctx, queryParams)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "users_email_unique_idx" {
		return model.User{}, userDomain.ErrEmailTaken
	}
	if err != nil {
		return model.User{}, fmt.Errorf("create user: %w", err)
	}
	return mapManagedUser(row), nil
}

func (r *Repository) Update(ctx context.Context, input userDomain.UpdateInput) (model.User, error) {
	// Parse the user identifier before updating PostgreSQL.
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return model.User{}, userDomain.ErrInvalidInput
	}
	// Update the managed user record.
	queryParams := db.UpdateUserParams{
		ID:       id,
		Email:    input.Email,
		FullName: input.FullName,
		Phone:    input.Phone,
		Status:   input.Status,
	}
	row, err := r.queries.UpdateUser(ctx, queryParams)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, userDomain.ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("update user: %w", err)
	}
	return mapUpdatedUser(row), nil
}

func (r *Repository) Delete(ctx context.Context, rawID string) error {
	// Parse the user identifier before deleting from PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return userDomain.ErrInvalidInput
	}
	affected, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if affected == 0 {
		return userDomain.ErrNotFound
	}
	return nil
}

func mapUser(row db.FindUserByIDRow) model.User {
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

func mapListUser(row db.ListUsersRow) model.User {
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

func mapManagedUser(row db.CreateManagedUserRow) model.User {
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

func mapUpdatedUser(row db.UpdateUserRow) model.User {
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
