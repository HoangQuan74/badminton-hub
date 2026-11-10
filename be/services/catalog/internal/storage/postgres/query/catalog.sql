-- The caller verifies tenant/venue scope with Identity before using these queries.
-- name: ListCourtsByVenue :many
SELECT id, venue_id, organization_id, name, status, slot_interval_minutes, created_at, updated_at
FROM catalog.courts
WHERE venue_id = sqlc.arg(venue_id) AND organization_id = sqlc.arg(organization_id)
  AND (NOT sqlc.arg(active_only)::boolean OR status = 'ACTIVE')
ORDER BY name;

-- name: GetCourt :one
SELECT id, venue_id, organization_id, name, status, slot_interval_minutes, created_at, updated_at
FROM catalog.courts WHERE id = $1;

-- name: CreateCourt :one
INSERT INTO catalog.courts (id, organization_id, venue_id, name, status, slot_interval_minutes)
VALUES (sqlc.arg(id), sqlc.arg(organization_id), sqlc.arg(venue_id), sqlc.arg(name), 'ACTIVE', sqlc.arg(slot_interval_minutes))
RETURNING id, venue_id, organization_id, name, status, slot_interval_minutes, created_at, updated_at;

-- name: UpdateCourt :one
UPDATE catalog.courts SET name = sqlc.arg(name), slot_interval_minutes = sqlc.arg(slot_interval_minutes), status = sqlc.arg(status), updated_at = NOW()
WHERE id = sqlc.arg(id) AND venue_id = sqlc.arg(venue_id) AND organization_id = sqlc.arg(organization_id)
RETURNING id, venue_id, organization_id, name, status, slot_interval_minutes, created_at, updated_at;

-- name: ArchiveCourt :one
UPDATE catalog.courts SET status = 'INACTIVE', updated_at = NOW()
WHERE id = sqlc.arg(id) AND venue_id = sqlc.arg(venue_id) AND organization_id = sqlc.arg(organization_id)
RETURNING id, venue_id, organization_id, name, status, slot_interval_minutes, created_at, updated_at;

-- name: CreateVenueImage :one
INSERT INTO catalog.venue_images (id, venue_id, object_key, content_type, size_bytes)
VALUES (sqlc.arg(id), sqlc.arg(venue_id), sqlc.arg(object_key), sqlc.arg(content_type), sqlc.arg(size_bytes))
RETURNING id, venue_id, object_key, content_type, size_bytes, created_at;

-- name: ListVenueImages :many
SELECT id, venue_id, object_key, content_type, size_bytes, created_at
FROM catalog.venue_images WHERE venue_id = $1 ORDER BY created_at, id;

-- name: GetVenueImage :one
SELECT id, venue_id, object_key, content_type, size_bytes, created_at
FROM catalog.venue_images WHERE id = $1;

-- name: IdentityCutoverReady :one
SELECT ready FROM catalog.identity_cutover_state WHERE singleton = TRUE;
