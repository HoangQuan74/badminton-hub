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

func (r *Repository) List(ctx context.Context, query model.ListQuery) ([]model.Staff, error) {
	// Load the requested staff page from PostgreSQL.
	queryParams := db.ListStaffParams{
		LimitCount:  int32(query.Limit),
		OffsetCount: int32(query.Offset),
		Search:      query.Search,
		CreatedFrom: postgresTransaction.TimestampPointer(query.CreatedFrom),
		CreatedTo:   postgresTransaction.TimestampPointer(query.CreatedTo),
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

func (r *Repository) FindByID(ctx context.Context, rawID string) (model.Staff, error) {
	// Parse the staff identifier before querying PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.Staff{}, staffDomain.ErrInvalidInput
	}
	row, err := r.queries.FindStaffByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Staff{}, staffDomain.ErrNotFound
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

func (r *Repository) Create(ctx context.Context, input staffDomain.CreateParams) (model.Staff, error) {
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

func (r *Repository) Update(ctx context.Context, input staffDomain.UpdateInput) (model.Staff, error) {
	// Parse the staff identifier before updating PostgreSQL.
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return model.Staff{}, staffDomain.ErrInvalidInput
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
		return model.Staff{}, staffDomain.ErrNotFound
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

func (r *Repository) UpdateAccess(ctx context.Context, input staffDomain.AccessInput) (model.Staff, error) {
	// Parse the staff identifier before changing access.
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return model.Staff{}, staffDomain.ErrInvalidInput
	}
	// Update the staff access settings.
	queryParams := db.UpdateStaffAccessParams{
		ID:      id,
		IsAdmin: input.IsAdmin,
		RoleID:  input.RoleID,
	}
	row, err := r.queries.UpdateStaffAccess(ctx, queryParams)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Staff{}, staffDomain.ErrNotFound
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

func (r *Repository) Delete(ctx context.Context, rawID string) error {
	// Parse the staff identifier before deleting from PostgreSQL.
	id, err := uuid.Parse(rawID)
	if err != nil {
		return staffDomain.ErrInvalidInput
	}
	affected, err := r.queries.DeleteStaff(ctx, id)
	if err != nil {
		return fmt.Errorf("delete staff: %w", err)
	}
	if affected == 0 {
		return staffDomain.ErrNotFound
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
