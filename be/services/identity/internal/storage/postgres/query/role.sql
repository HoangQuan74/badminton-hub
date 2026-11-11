-- name: ListRoles :many
SELECT id, name, description, created_at FROM identity.roles ORDER BY name;

-- name: CreateRole :one
INSERT INTO identity.roles (id, name, description) VALUES ($1, $2, $3)
RETURNING id, name, description, created_at;

-- name: GetRole :one
SELECT id, name, description, created_at FROM identity.roles WHERE id = $1;

-- name: UpdateRole :one
UPDATE identity.roles SET name = $2, description = $3, updated_at = NOW() WHERE id = $1
RETURNING id, name, description, created_at;

-- name: DeleteRole :execrows
DELETE FROM identity.roles WHERE id = $1;

-- name: DeleteRolePermissions :exec
DELETE FROM identity.role_permissions WHERE role_id = $1;

-- name: AddRolePermission :exec
INSERT INTO identity.role_permissions (role_id, permission_id) VALUES ($1, $2);

-- name: ListPermissions :many
SELECT id, module, path, action, description FROM identity.permissions ORDER BY module, path, action;
