package rolepostgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/role"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
	postgresTransaction "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/transaction"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: db.New(pool)}
}

func makeRole(id uuid.UUID, name, description string, created time.Time) model.Role {
	return model.Role{ID: id.String(), Name: name, Description: description, CreatedAt: created}
}

func (r *Repository) ListRoles(ctx context.Context) ([]model.Role, error) {
	rows, err := r.queries.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]model.Role, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, makeRole(row.ID, row.Name, row.Description, row.CreatedAt.Time))
	}
	return roles, nil
}

func (r *Repository) CreateRole(ctx context.Context, name, description string) (model.Role, error) {
	row, err := r.queries.CreateRole(ctx, db.CreateRoleParams{ID: uuid.New(), Name: name, Description: description})
	if err != nil {
		return model.Role{}, err
	}
	return makeRole(row.ID, row.Name, row.Description, row.CreatedAt.Time), nil
}

func (r *Repository) GetRole(ctx context.Context, rawID string) (model.Role, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.Role{}, role.ErrInvalidInput
	}
	row, err := r.queries.GetRole(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Role{}, role.ErrNotFound
	}
	if err != nil {
		return model.Role{}, err
	}
	return makeRole(row.ID, row.Name, row.Description, row.CreatedAt.Time), nil
}

func (r *Repository) UpdateRole(ctx context.Context, rawID, name, description string) (model.Role, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.Role{}, role.ErrInvalidInput
	}
	row, err := r.queries.UpdateRole(ctx, db.UpdateRoleParams{ID: id, Name: name, Description: description})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Role{}, role.ErrNotFound
	}
	if err != nil {
		return model.Role{}, err
	}
	return makeRole(row.ID, row.Name, row.Description, row.CreatedAt.Time), nil
}

func (r *Repository) DeleteRole(ctx context.Context, rawID string) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return role.ErrInvalidInput
	}
	count, err := r.queries.DeleteRole(ctx, id)
	if err != nil {
		return err
	}
	if count == 0 {
		return role.ErrNotFound
	}
	return nil
}

func (r *Repository) ReplaceRolePermissions(ctx context.Context, rawID string, rawPermissions []string) (model.Role, error) {
	roleID, err := uuid.Parse(rawID)
	if err != nil {
		return model.Role{}, role.ErrInvalidInput
	}
	permissions := make([]uuid.UUID, 0, len(rawPermissions))
	for _, raw := range rawPermissions {
		id, err := uuid.Parse(raw)
		if err != nil {
			return model.Role{}, role.ErrInvalidInput
		}
		permissions = append(permissions, id)
	}
	var role model.Role
	err = postgresTransaction.WithinTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.queries.WithTx(tx)
		if err := queries.DeleteRolePermissions(ctx, roleID); err != nil {
			return err
		}
		for _, id := range permissions {
			if err := queries.AddRolePermission(ctx, db.AddRolePermissionParams{RoleID: roleID, PermissionID: id}); err != nil {
				return err
			}
		}
		row, err := queries.GetRole(ctx, roleID)
		if err != nil {
			return err
		}
		role = makeRole(row.ID, row.Name, row.Description, row.CreatedAt.Time)
		return nil
	})
	if err != nil {
		return model.Role{}, err
	}
	return role, nil
}

func (r *Repository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	rows, err := r.queries.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	permissions := make([]model.Permission, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, model.Permission{ID: row.ID.String(), Module: row.Module, Path: row.Path, Action: row.Action, Description: row.Description})
	}
	return permissions, nil
}
