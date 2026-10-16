-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, full_name, phone)
VALUES ($1, $2, $3, $4, NULLIF(sqlc.arg(phone)::text, ''))
RETURNING id, email, password_hash, full_name, COALESCE(phone, '') AS phone, status, created_at, updated_at;

-- name: FindUserByEmail :one
SELECT id, email, password_hash, full_name, COALESCE(phone, '') AS phone, status, created_at, updated_at
FROM users
WHERE LOWER(email) = LOWER($1);

-- name: FindUserByID :one
SELECT id, email, password_hash, full_name, COALESCE(phone, '') AS phone, status, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateAuthSession :exec
INSERT INTO auth_sessions (id, user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: FindActiveSessionForUpdate :one
SELECT s.user_id
FROM auth_sessions s
JOIN users u ON u.id = s.user_id
WHERE s.refresh_token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > sqlc.arg(now)::timestamptz
  AND u.status = 'active'
FOR UPDATE OF s;

-- name: RevokeAuthSession :execrows
UPDATE auth_sessions
SET revoked_at = $2
WHERE refresh_token_hash = $1
  AND revoked_at IS NULL;
