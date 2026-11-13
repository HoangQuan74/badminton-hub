-- name: ListPermissions :many
SELECT id, module, path, action, description FROM identity.permissions ORDER BY module, path, action;

-- name: CreatePermission :one
INSERT INTO identity.permissions (id, module, path, action, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, module, path, action, description;

-- name: GetPermission :one
SELECT id, module, path, action, description FROM identity.permissions WHERE id = $1;

-- name: UpdatePermission :one
UPDATE identity.permissions SET module = $2, path = $3, action = $4, description = $5 WHERE id = $1
RETURNING id, module, path, action, description;

-- name: DeletePermission :execrows
DELETE FROM identity.permissions WHERE id = $1;
