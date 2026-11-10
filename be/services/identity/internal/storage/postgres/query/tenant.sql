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

-- name: ListIdentityPublicVenues :many
SELECT v.id, v.organization_id, v.name, v.slug, v.address, v.timezone, v.status
FROM identity.venues AS v JOIN identity.organizations AS o ON o.id = v.organization_id
WHERE v.status = 'PUBLISHED' AND o.status = 'ACTIVE' ORDER BY v.name;

-- name: GetIdentityPublicVenueBySlug :one
SELECT v.id, v.organization_id, v.name, v.slug, v.address, v.timezone, v.status
FROM identity.venues AS v JOIN identity.organizations AS o ON o.id = v.organization_id
WHERE LOWER(v.slug) = LOWER(sqlc.arg(slug)) AND v.status = 'PUBLISHED' AND o.status = 'ACTIVE';

-- name: GetIdentityPublicVenueByID :one
SELECT v.id, v.organization_id, v.name, v.slug, v.address, v.timezone, v.status
FROM identity.venues AS v JOIN identity.organizations AS o ON o.id = v.organization_id
WHERE v.id = sqlc.arg(venue_id) AND v.status = 'PUBLISHED' AND o.status = 'ACTIVE';

-- name: ListIdentityTenantVenues :many
SELECT v.id, v.organization_id, v.name, v.slug, v.address, v.timezone, v.status
FROM identity.venues AS v
JOIN identity.organizations AS o ON o.id = v.organization_id
JOIN identity.organization_members AS m ON m.organization_id = v.organization_id
WHERE m.user_id = sqlc.arg(actor_id) AND m.status = 'ACTIVE' AND o.status <> 'ARCHIVED'
  AND v.status <> 'ARCHIVED' AND (
    m.role IN ('tenant_owner', 'tenant_staff') OR (m.role = 'venue_staff' AND EXISTS (
      SELECT 1 FROM identity.venue_staff AS a WHERE a.venue_id = v.id AND a.user_id = m.user_id AND a.status = 'ACTIVE'
    ))
  )
ORDER BY v.name;

-- name: GetIdentityAuthorizedVenue :one
SELECT v.id, v.organization_id, v.name, v.slug, v.address, v.timezone, v.status
FROM identity.venues AS v
JOIN identity.organizations AS o ON o.id = v.organization_id
JOIN identity.organization_members AS m ON m.organization_id = v.organization_id
WHERE v.id = sqlc.arg(venue_id) AND m.user_id = sqlc.arg(actor_id)
  AND m.status = 'ACTIVE' AND o.status <> 'ARCHIVED' AND v.status <> 'ARCHIVED'
  AND (
    m.role = 'tenant_owner' OR (NOT sqlc.arg(require_owner)::boolean AND (
      m.role = 'tenant_staff' OR (m.role = 'venue_staff' AND EXISTS (
        SELECT 1 FROM identity.venue_staff AS a WHERE a.venue_id = v.id AND a.user_id = m.user_id AND a.status = 'ACTIVE'
      ))
    ))
  );

-- name: CreateIdentityVenue :one
INSERT INTO identity.venues (id, organization_id, name, slug, address, timezone, status)
SELECT sqlc.arg(id), o.id, sqlc.arg(name), sqlc.arg(slug), sqlc.arg(address), sqlc.arg(timezone), 'DRAFT'
FROM identity.organizations AS o
JOIN identity.organization_members AS m ON m.organization_id = o.id
WHERE o.id = sqlc.arg(organization_id) AND o.status = 'ACTIVE'
  AND m.user_id = sqlc.arg(actor_id) AND m.status = 'ACTIVE' AND m.role = 'tenant_owner'
RETURNING id, organization_id, name, slug, address, timezone, status;

-- name: UpdateIdentityVenue :one
UPDATE identity.venues AS v SET
  name = COALESCE(sqlc.narg(name)::text, v.name), slug = COALESCE(sqlc.narg(slug)::text, v.slug),
  address = COALESCE(sqlc.narg(address)::text, v.address), timezone = COALESCE(sqlc.narg(timezone)::text, v.timezone), updated_at = NOW()
WHERE v.id = sqlc.arg(venue_id) AND v.status <> 'ARCHIVED' AND EXISTS (
  SELECT 1 FROM identity.organization_members AS m
  WHERE m.organization_id = v.organization_id AND m.user_id = sqlc.arg(actor_id)
    AND m.status = 'ACTIVE' AND m.role = 'tenant_owner'
)
RETURNING id, organization_id, name, slug, address, timezone, status;

-- name: ArchiveIdentityVenue :one
UPDATE identity.venues AS v SET status = 'ARCHIVED', updated_at = NOW()
WHERE v.id = sqlc.arg(venue_id) AND v.status <> 'ARCHIVED' AND EXISTS (
  SELECT 1 FROM identity.organization_members AS m
  WHERE m.organization_id = v.organization_id AND m.user_id = sqlc.arg(actor_id)
    AND m.status = 'ACTIVE' AND m.role = 'tenant_owner'
)
RETURNING id, organization_id, name, slug, address, timezone, status;

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
