CREATE TABLE catalog.venue_images (
    id UUID PRIMARY KEY,
    venue_id UUID NOT NULL REFERENCES catalog.venues(id),
    object_key TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL CHECK (content_type IN ('image/jpeg', 'image/png')),
    size_bytes INTEGER NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 2097152),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX venue_images_venue_created_idx ON catalog.venue_images (venue_id, created_at, id);
