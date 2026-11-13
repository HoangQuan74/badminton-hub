package permissionpostgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/permission"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

type Repository struct{ queries *db.Queries }

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: db.New(pool)}
}

func (r *Repository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	rows, err := r.queries.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	permissions := make([]model.Permission, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, permissionModel(
			row.ID,
			row.Module,
			row.Path,
			row.Action,
			row.Description,
		))
	}

	return permissions, nil
}

func (r *Repository) CreatePermission(ctx context.Context, input permission.Input) (model.Permission, error) {
	row, err := r.queries.CreatePermission(ctx, db.CreatePermissionParams{
		ID:          uuid.New(),
		Module:      input.Module,
		Path:        input.Path,
		Action:      input.Action,
		Description: input.Description,
	})
	if err != nil {
		return model.Permission{}, identityException.PermissionPersistenceError(err)
	}

	return permissionModel(
		row.ID,
		row.Module,
		row.Path,
		row.Action,
		row.Description,
	), nil
}

func (r *Repository) GetPermission(ctx context.Context, rawID string) (model.Permission, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.Permission{}, identityException.ErrPermissionInvalidInput
	}

	row, err := r.queries.GetPermission(ctx, id)
	if err != nil {
		return model.Permission{}, identityException.PermissionPersistenceError(err)
	}

	return permissionModel(
		row.ID,
		row.Module,
		row.Path,
		row.Action,
		row.Description,
	), nil
}

func (r *Repository) UpdatePermission(ctx context.Context, rawID string, input permission.Input) (model.Permission, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return model.Permission{}, identityException.ErrPermissionInvalidInput
	}

	row, err := r.queries.UpdatePermission(ctx, db.UpdatePermissionParams{
		ID:          id,
		Module:      input.Module,
		Path:        input.Path,
		Action:      input.Action,
		Description: input.Description,
	})
	if err != nil {
		return model.Permission{}, identityException.PermissionPersistenceError(err)
	}

	return permissionModel(
		row.ID,
		row.Module,
		row.Path,
		row.Action,
		row.Description,
	), nil
}

func (r *Repository) DeletePermission(ctx context.Context, rawID string) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return identityException.ErrPermissionInvalidInput
	}

	count, err := r.queries.DeletePermission(ctx, id)
	if err != nil {
		return err
	}

	if count == 0 {
		return identityException.ErrPermissionNotFound
	}

	return nil
}

func permissionModel(id uuid.UUID, module, path, action, description string) model.Permission {
	return model.Permission{
		ID:          id.String(),
		Module:      module,
		Path:        path,
		Action:      action,
		Description: description,
	}
}
