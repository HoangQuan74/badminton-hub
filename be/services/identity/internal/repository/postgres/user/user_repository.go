package userpostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	postgresHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/helper"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *Repository) ListUsers(ctx context.Context, query model.ListQuery) ([]model.User, error) {
	// Load the requested user page from PostgreSQL.
	queryParams := db.ListUsersParams{
		LimitCount:  int32(query.Limit),
		OffsetCount: int32(query.Offset),
		Search:      query.Search,
		CreatedFrom: postgresHelper.MapOptionalTimestamp(query.CreatedFrom),
		CreatedTo:   postgresHelper.MapOptionalTimestamp(query.CreatedTo),
	}
	rows, err := r.queries.ListUsers(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	// Map storage rows to reusable domain models.
	users := make([]model.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, postgresHelper.MapUser(db.CreateUserRow(row)))
	}

	return users, nil
}

func (r *Repository) GetUser(ctx context.Context, rawID string) (model.User, error) {
	// Parse the user identifier before querying PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.User{}, identityException.ErrUserInvalidInput
	}

	row, err := r.queries.FindUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, identityException.ErrUserNotFound
	}

	if err != nil {
		return model.User{}, fmt.Errorf("find user: %w", err)
	}

	return postgresHelper.MapUserByID(row), nil
}

func (r *Repository) CreateUser(ctx context.Context, input userDomain.CreateParams) (model.User, error) {
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
	if identityException.IsUserEmailConflict(err) {
		return model.User{}, identityException.ErrUserEmailTaken
	}

	if err != nil {
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return postgresHelper.MapUser(db.CreateUserRow(row)), nil
}

func (r *Repository) UpdateUser(ctx context.Context, input userDomain.UpdateInput) (model.User, error) {
	// Parse the user identifier before updating PostgreSQL.
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return model.User{}, identityException.ErrUserInvalidInput
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
		return model.User{}, identityException.ErrUserNotFound
	}

	if err != nil {
		return model.User{}, fmt.Errorf("update user: %w", err)
	}

	return postgresHelper.MapUser(db.CreateUserRow(row)), nil
}

func (r *Repository) DeleteUser(ctx context.Context, rawID string) error {
	// Parse the user identifier before deleting from PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return identityException.ErrUserInvalidInput
	}

	affected, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if affected == 0 {
		return identityException.ErrUserNotFound
	}

	return nil
}
