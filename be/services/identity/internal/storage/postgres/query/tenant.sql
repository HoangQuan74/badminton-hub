-- name: ListTenantOrganizations :many
SELECT o.id, o.name, o.slug, o.status, o.created_at, o.updated_at
FROM identity.organizations AS o
JOIN identity.organization_members AS m ON m.organization_id = o.id
WHERE m.user_id = sqlc.arg(actor_id) AND m.status = 'ACTIVE' AND o.status <> 'ARCHIVED'
ORDER BY o.name;

-- name: GetTenantOrganization :one
SELECT o.id, o.name, o.slug, o.status, o.created_at, o.updated_at
FROM identity.organizations AS o
JOIN identity.organization_members AS m ON m.organization_id = o.id
WHERE o.id = sqlc.arg(organization_id) AND m.user_id = sqlc.arg(actor_id)
  AND m.status = 'ACTIVE' AND o.status <> 'ARCHIVED';

-- name: FindOrganizationMembership :one
SELECT m.role FROM identity.organization_members AS m
JOIN identity.organizations AS o ON o.id = m.organization_id
WHERE m.organization_id = sqlc.arg(organization_id) AND m.user_id = sqlc.arg(actor_id)
  AND m.status = 'ACTIVE' AND o.status <> 'ARCHIVED';

-- name: CreateOrganization :one
WITH new_organization AS (
  INSERT INTO identity.organizations (id, name, slug, status)
  VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(slug), 'PENDING_VERIFICATION')
  RETURNING id, name, slug, status, created_at, updated_at
), new_membership AS (
  INSERT INTO identity.organization_members (organization_id, user_id, role, status)
  SELECT id, sqlc.arg(actor_id), 'tenant_owner', 'ACTIVE' FROM new_organization
  RETURNING organization_id
)
SELECT o.id, o.name, o.slug, o.status, o.created_at, o.updated_at
FROM new_organization AS o JOIN new_membership AS m ON m.organization_id = o.id;

-- name: UpdateOrganization :one
UPDATE identity.organizations AS o SET
  name = COALESCE(sqlc.narg(name)::text, o.name),
  slug = COALESCE(sqlc.narg(slug)::text, o.slug), updated_at = NOW()
WHERE o.id = sqlc.arg(organization_id) AND o.status IN ('PENDING_VERIFICATION', 'ACTIVE') AND EXISTS (
  SELECT 1 FROM identity.organization_members AS m
  WHERE m.organization_id = o.id AND m.user_id = sqlc.arg(actor_id)
    AND m.status = 'ACTIVE' AND m.role = 'tenant_owner'
)
RETURNING id, name, slug, status, created_at, updated_at;

-- name: ArchivePendingOrganization :one
UPDATE identity.organizations AS o SET status = 'ARCHIVED', updated_at = NOW()
WHERE o.id = sqlc.arg(organization_id) AND o.status = 'PENDING_VERIFICATION' AND EXISTS (
  SELECT 1 FROM identity.organization_members AS m
  WHERE m.organization_id = o.id AND m.user_id = sqlc.arg(actor_id)
    AND m.status = 'ACTIVE' AND m.role = 'tenant_owner'
)
RETURNING id, name, slug, status, created_at, updated_at;

-- name: LockTenantCreator :one
SELECT u.id, u.email, u.full_name, COALESCE(u.phone, '')::text AS phone,
       u.status, u.is_admin, COALESCE(r.name, '')::text AS role_name,
       u.created_at, u.updated_at
FROM identity.users AS u
LEFT JOIN identity.roles AS r ON r.id = u.role_id
WHERE u.id = $1
FOR UPDATE OF u;

-- name: FindTenantCreateReplay :one
SELECT request_hash, response, expires_at
FROM identity.onboard_requests
WHERE actor_id = $1 AND idempotency_key = $2 AND expires_at > clock_timestamp();

-- name: DeleteExpiredTenantCreateKey :exec
DELETE FROM identity.onboard_requests
WHERE actor_id = $1 AND idempotency_key = $2 AND expires_at <= clock_timestamp();

-- name: SetTenantOwnerRole :execrows
UPDATE identity.users AS u
SET role_id = (SELECT r.id FROM identity.roles AS r WHERE LOWER(r.name) = 'tenant_owner'), updated_at = NOW()
WHERE u.id = $1 AND EXISTS (SELECT 1 FROM identity.roles AS r WHERE LOWER(r.name) = 'tenant_owner');

-- name: SaveTenantCreateReplay :one
INSERT INTO identity.onboard_requests (actor_id, idempotency_key, request_hash, response, expires_at)
VALUES ($1, $2, $3, $4, clock_timestamp() + INTERVAL '24 hours')
RETURNING expires_at;

-- name: SaveTenantRegistration :exec
INSERT INTO identity.tenant_registrations (organization_id, contact_name, contact_email, contact_phone, note)
VALUES ($1, $2, $3, $4, $5);

-- name: ListTenantRegistrations :many
SELECT o.id, o.name, o.slug, o.status, o.created_at, o.updated_at,
       r.contact_name, r.contact_email, r.contact_phone, r.note,
       COALESCE(v.decision, '')::text AS review_decision, COALESCE(v.note, '')::text AS review_note,
       COALESCE(v.reviewer_id::text, '')::text AS reviewer_id, v.created_at AS reviewed_at
FROM identity.organizations o
JOIN identity.tenant_registrations r ON r.organization_id = o.id
LEFT JOIN identity.tenant_registration_reviews v ON v.organization_id = o.id
WHERE sqlc.arg(status_filter)::text = '' OR o.status = sqlc.arg(status_filter)
ORDER BY o.created_at DESC;

-- name: GetTenantRegistration :one
SELECT o.id, o.name, o.slug, o.status, o.created_at, o.updated_at,
       r.contact_name, r.contact_email, r.contact_phone, r.note,
       COALESCE(v.decision, '')::text AS review_decision, COALESCE(v.note, '')::text AS review_note,
       COALESCE(v.reviewer_id::text, '')::text AS reviewer_id, v.created_at AS reviewed_at
FROM identity.organizations o
JOIN identity.tenant_registrations r ON r.organization_id = o.id
LEFT JOIN identity.tenant_registration_reviews v ON v.organization_id = o.id
WHERE o.id = $1;

-- name: LockTenantRegistration :one
SELECT o.status FROM identity.organizations o
JOIN identity.tenant_registrations r ON r.organization_id = o.id
WHERE o.id = $1 FOR UPDATE OF o;

-- name: ApplyTenantRegistrationDecision :exec
UPDATE identity.organizations SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: RecordTenantRegistrationReview :exec
INSERT INTO identity.tenant_registration_reviews (organization_id, reviewer_id, decision, note)
VALUES ($1, $2, $3, $4);
