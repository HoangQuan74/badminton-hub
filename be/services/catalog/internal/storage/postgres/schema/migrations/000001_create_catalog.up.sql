CREATE SCHEMA IF NOT EXISTS catalog;

CREATE TABLE catalog.organizations (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('PENDING_VERIFICATION', 'ACTIVE', 'SUSPENDED', 'REJECTED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX organizations_slug_unique_idx ON catalog.organizations (LOWER(slug));

CREATE TABLE catalog.organization_members (
    organization_id UUID NOT NULL REFERENCES catalog.organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('tenant_owner', 'tenant_staff')),
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

CREATE TABLE catalog.venues (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES catalog.organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'Asia/Ho_Chi_Minh',
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'PUBLISHED', 'SUSPENDED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX venues_slug_unique_idx ON catalog.venues (LOWER(slug));
CREATE INDEX venues_organization_id_idx ON catalog.venues (organization_id);

CREATE TABLE catalog.courts (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES catalog.organizations(id) ON DELETE CASCADE,
    venue_id UUID NOT NULL REFERENCES catalog.venues(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    slot_interval_minutes INTEGER NOT NULL DEFAULT 30 CHECK (slot_interval_minutes BETWEEN 15 AND 180 AND slot_interval_minutes % 15 = 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX courts_venue_id_idx ON catalog.courts (venue_id);
CREATE UNIQUE INDEX courts_venue_name_unique_idx ON catalog.courts (venue_id, LOWER(name));
