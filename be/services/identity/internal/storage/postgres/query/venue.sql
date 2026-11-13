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

-- name: GetVenueOrganizationAccess :one
SELECT o.status,m.role
FROM identity.organizations o
JOIN identity.organization_members m ON m.organization_id=o.id
WHERE o.id=sqlc.arg(organization_id) AND m.user_id=sqlc.arg(actor_id)
 AND m.status='ACTIVE' AND o.status<>'ARCHIVED';
