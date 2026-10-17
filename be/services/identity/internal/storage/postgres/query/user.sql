-- name: ListUsers :many
SELECT id, email, password_hash, full_name, COALESCE(phone, '') AS phone, status, created_at, updated_at
FROM identity.users
WHERE (sqlc.arg(search)::text = ''
       OR email ILIKE '%' || sqlc.arg(search)::text || '%'
       OR full_name ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_to)::timestamptz IS NULL OR created_at <= sqlc.narg(created_to)::timestamptz)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count)::int OFFSET sqlc.arg(offset_count)::int;

-- name: CreateManagedUser :one
INSERT INTO identity.users (id, email, password_hash, full_name, phone, status)
VALUES ($1, $2, $3, $4, NULLIF(sqlc.arg(phone)::text, ''), $5)
RETURNING id, email, password_hash, full_name, COALESCE(phone, '') AS phone, status, created_at, updated_at;

-- name: UpdateUser :one
UPDATE identity.users
SET email = $2,
    full_name = $3,
    phone = NULLIF(sqlc.arg(phone)::text, ''),
    status = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING id, email, password_hash, full_name, COALESCE(phone, '') AS phone, status, created_at, updated_at;

-- name: DeleteUser :execrows
DELETE FROM identity.users WHERE id = $1;
