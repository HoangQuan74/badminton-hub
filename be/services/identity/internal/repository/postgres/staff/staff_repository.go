package staffpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
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

func (r *Repository) ListStaff(ctx context.Context, query model.ListQuery) ([]model.Staff, error) {
	// Load the requested staff page from PostgreSQL.
	queryParams := db.ListStaffParams{
		LimitCount:  int32(query.Limit),
		OffsetCount: int32(query.Offset),
		Search:      query.Search,
		CreatedFrom: postgresHelper.MapOptionalTimestamp(query.CreatedFrom),
		CreatedTo:   postgresHelper.MapOptionalTimestamp(query.CreatedTo),
	}
	rows, err := r.queries.ListStaff(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("list staff: %w", err)
	}

	// Map storage rows to reusable domain models.
	staffMembers := make([]model.Staff, 0, len(rows))
	for _, row := range rows {
		staffMembers = append(staffMembers, mapStaffValues(
			row.ID,
			row.Email,
			row.FullName,
			row.IsAdmin,
			row.RoleID,
			row.RoleName,
			row.Status,
			row.CreatedAt.Time,
		))
	}

	return staffMembers, nil
}

func (r *Repository) GetStaff(ctx context.Context, rawID string) (model.Staff, error) {
	// Parse the staff identifier before querying PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	row, err := r.queries.FindStaffByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Staff{}, identityException.ErrStaffNotFound
	}

	if err != nil {
		return model.Staff{}, fmt.Errorf("find staff: %w", err)
	}

	return mapStaffValues(
		row.ID,
		row.Email,
		row.FullName,
		row.IsAdmin,
		row.RoleID,
		row.RoleName,
		row.Status,
		row.CreatedAt.Time,
	), nil
}

func (r *Repository) CreateStaff(ctx context.Context, input staffDomain.CreateStaffParams) (model.Staff, error) {
	// Create a staff record.
	queryParams := db.CreateStaffParams{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		FullName:     input.FullName,
		IsAdmin:      input.IsAdmin,
		RoleID:       input.RoleID,
	}
	row, err := r.queries.CreateStaff(ctx, queryParams)
	if err != nil {
		return model.Staff{}, fmt.Errorf("create staff: %w", err)
	}

	return mapStaffValues(
		row.ID,
		row.Email,
		row.FullName,
		row.IsAdmin,
		row.RoleID,
		"",
		row.Status,
		row.CreatedAt.Time,
	), nil
}

func (r *Repository) UpdateStaff(ctx context.Context, input staffDomain.UpdateStaffInput) (model.Staff, error) {
	// Parse the staff identifier before updating PostgreSQL.
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	// Update the staff record.
	queryParams := db.UpdateStaffParams{
		ID:       id,
		Email:    input.Email,
		FullName: input.FullName,
		Status:   input.Status,
	}
	row, err := r.queries.UpdateStaff(ctx, queryParams)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Staff{}, identityException.ErrStaffNotFound
	}

	if err != nil {
		return model.Staff{}, fmt.Errorf("update staff: %w", err)
	}

	return mapStaffValues(
		row.ID,
		row.Email,
		row.FullName,
		row.IsAdmin,
		row.RoleID,
		"",
		row.Status,
		row.CreatedAt.Time,
	), nil
}

func (r *Repository) UpdateStaffAccess(ctx context.Context, input staffDomain.UpdateStaffAccessInput) (model.Staff, error) {
	// Parse the staff identifier before changing access.
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return model.Staff{}, identityException.ErrStaffInvalidInput
	}

	// Update the staff access settings.
	queryParams := db.UpdateStaffAccessParams{
		ID:      id,
		IsAdmin: input.IsAdmin,
		RoleID:  input.RoleID,
	}
	row, err := r.queries.UpdateStaffAccess(ctx, queryParams)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Staff{}, identityException.ErrStaffNotFound
	}

	if err != nil {
		return model.Staff{}, fmt.Errorf("update staff access: %w", err)
	}

	return mapStaffValues(
		row.ID,
		row.Email,
		row.FullName,
		row.IsAdmin,
		row.RoleID,
		"",
		row.Status,
		row.CreatedAt.Time,
	), nil
}

func (r *Repository) DeleteStaff(ctx context.Context, rawID string) error {
	// Parse the staff identifier before deleting from PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return identityException.ErrStaffInvalidInput
	}

	affected, err := r.queries.DeleteStaff(ctx, id)
	if err != nil {
		return fmt.Errorf("delete staff: %w", err)
	}

	if affected == 0 {
		return identityException.ErrStaffNotFound
	}

	return nil
}

func mapStaffValues(id uuid.UUID, email, fullName string, isAdmin bool, rawRoleID any, roleName, status string, createdAt time.Time) model.Staff {
	return model.Staff{
		ID:        id.String(),
		Email:     email,
		FullName:  fullName,
		IsAdmin:   isAdmin,
		RoleID:    roleID(rawRoleID),
		RoleName:  roleName,
		Status:    status,
		CreatedAt: createdAt,
	}
}

func roleID(value any) string {
	if value == nil {
		return ""
	}

	if result, ok := value.(string); ok {
		return result
	}

	return fmt.Sprint(value)
}
