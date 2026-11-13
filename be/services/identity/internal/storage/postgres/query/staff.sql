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

-- name: SetIdentityOrganizationStaff :one
INSERT INTO identity.organization_members (organization_id, user_id, role, status)
SELECT o.id, u.id, sqlc.arg(role), sqlc.arg(status)
FROM identity.organizations AS o
JOIN identity.organization_members AS owner ON owner.organization_id = o.id
JOIN identity.users AS u ON u.id = sqlc.arg(user_id) AND u.status = 'active'
JOIN identity.roles AS target_role ON target_role.id = u.role_id AND target_role.name = sqlc.arg(role)
WHERE o.id = sqlc.arg(organization_id) AND o.status = 'ACTIVE'
  AND owner.user_id = sqlc.arg(actor_id) AND owner.status = 'ACTIVE' AND owner.role = 'tenant_owner'
ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role, status = EXCLUDED.status
WHERE identity.organization_members.role <> 'tenant_owner'
RETURNING organization_id, user_id, role, status;

-- name: SetIdentityVenueStaff :one
INSERT INTO identity.venue_staff (organization_id, venue_id, user_id, status)
SELECT v.organization_id, v.id, member.user_id, sqlc.arg(status)
FROM identity.venues AS v
JOIN identity.organization_members AS owner ON owner.organization_id = v.organization_id
JOIN identity.organization_members AS member ON member.organization_id = v.organization_id
WHERE v.id = sqlc.arg(venue_id) AND v.status <> 'ARCHIVED'
  AND owner.user_id = sqlc.arg(actor_id) AND owner.status = 'ACTIVE' AND owner.role = 'tenant_owner'
  AND member.user_id = sqlc.arg(user_id) AND member.status = 'ACTIVE' AND member.role = 'venue_staff'
ON CONFLICT (venue_id, user_id) DO UPDATE SET status = EXCLUDED.status
RETURNING organization_id, venue_id, user_id, status;

-- name: FindManagedOrganization :one
SELECT o.status FROM identity.organizations o
JOIN identity.organization_members m ON m.organization_id = o.id
WHERE o.id = sqlc.arg(organization_id) AND o.status <> 'ARCHIVED'
 AND m.user_id = sqlc.arg(actor_id) AND m.role = 'tenant_owner' AND m.status = 'ACTIVE';

-- name: ListOrganizationStaff :many
SELECT m.organization_id, m.user_id, m.role, m.status, u.email, u.full_name,
 ARRAY(SELECT a.venue_id::text FROM identity.venue_staff a JOIN identity.venues v ON v.id = a.venue_id
       WHERE a.organization_id = m.organization_id AND a.user_id = m.user_id AND a.status = 'ACTIVE' AND v.status <> 'ARCHIVED')::text[] AS venue_ids
FROM identity.organization_members m JOIN identity.users u ON u.id = m.user_id
WHERE m.organization_id = $1 ORDER BY u.full_name, u.email;

-- name: LockStaffCandidate :one
SELECT u.id, u.status, u.is_admin, COALESCE(r.name,'')::text AS role_name
FROM identity.users u LEFT JOIN identity.roles r ON r.id = u.role_id
WHERE LOWER(u.email) = LOWER($1) FOR UPDATE OF u;

-- name: AssignStaffRole :execrows
UPDATE identity.users AS u SET role_id = (SELECT r.id FROM identity.roles r WHERE r.name = sqlc.arg(role)), updated_at = NOW()
WHERE u.id = sqlc.arg(user_id) AND EXISTS (SELECT 1 FROM identity.roles r WHERE r.name = sqlc.arg(role));

-- name: DeactivateOrganizationStaff :one
UPDATE identity.organization_members m SET status = 'INACTIVE'
WHERE m.organization_id = sqlc.arg(organization_id) AND m.user_id = sqlc.arg(user_id) AND m.role <> 'tenant_owner'
AND EXISTS (SELECT 1 FROM identity.organization_members owner
 JOIN identity.organizations o ON o.id = owner.organization_id
 WHERE owner.organization_id = m.organization_id AND o.status = 'ACTIVE'
 AND owner.user_id = sqlc.arg(actor_id) AND owner.role = 'tenant_owner' AND owner.status = 'ACTIVE')
RETURNING m.organization_id,m.user_id,m.role,m.status;

-- name: DeactivateMemberVenues :exec
UPDATE identity.venue_staff SET status = 'INACTIVE' WHERE organization_id = $1 AND user_id = $2;
