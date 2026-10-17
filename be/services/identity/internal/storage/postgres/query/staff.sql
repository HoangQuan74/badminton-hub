-- name: ListStaff :many
SELECT u.id, u.email, u.full_name, u.is_admin, COALESCE(u.role_id::text, '') AS role_id,
       COALESCE(r.name, '') AS role_name, u.status, u.created_at
FROM identity.users u
LEFT JOIN identity.roles r ON r.id = u.role_id
WHERE (u.is_admin = TRUE OR u.role_id IS NOT NULL)
  AND (sqlc.arg(search)::text = ''
       OR u.email ILIKE '%' || sqlc.arg(search)::text || '%'
       OR u.full_name ILIKE '%' || sqlc.arg(search)::text || '%'
       OR r.name ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR u.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz IS NULL OR u.created_at <= sqlc.narg(created_to)::timestamptz)
ORDER BY u.created_at DESC
LIMIT sqlc.arg(limit_count)::int OFFSET sqlc.arg(offset_count)::int;

-- name: FindStaffByID :one
SELECT u.id, u.email, u.full_name, u.is_admin, COALESCE(u.role_id::text, '') AS role_id,
       COALESCE(r.name, '') AS role_name, u.status, u.created_at
FROM identity.users u
LEFT JOIN identity.roles r ON r.id = u.role_id
WHERE u.id = $1 AND (u.is_admin = TRUE OR u.role_id IS NOT NULL);

-- name: CreateStaff :one
INSERT INTO identity.users (id, email, password_hash, full_name, is_admin, role_id, status)
VALUES ($1, $2, $3, $4, $5, NULLIF(sqlc.arg(role_id)::text, '')::uuid, 'active')
RETURNING id, email, full_name, is_admin, COALESCE(role_id::text, '') AS role_id, status, created_at;

-- name: UpdateStaff :one
UPDATE identity.users
SET email = $2, full_name = $3, status = $4, updated_at = NOW()
WHERE id = $1 AND (is_admin = TRUE OR role_id IS NOT NULL)
RETURNING id, email, full_name, is_admin, COALESCE(role_id::text, '') AS role_id, status, created_at;

-- name: UpdateStaffAccess :one
UPDATE identity.users
SET is_admin = $2, role_id = NULLIF(sqlc.arg(role_id)::text, '')::uuid, updated_at = NOW()
WHERE id = $1
RETURNING id, email, full_name, is_admin, COALESCE(role_id::text, '') AS role_id, status, created_at;

-- name: DeleteStaff :execrows
DELETE FROM identity.users
WHERE id = $1 AND (is_admin = TRUE OR role_id IS NOT NULL);
