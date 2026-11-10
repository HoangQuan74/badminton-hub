-- Identity is authoritative for tenants, venues, and staff access. The legacy
-- Catalog tables remain untouched until an explicit, verified cutover.
-- Provision the role in every environment; development seeds are optional.
INSERT INTO identity.roles (id, name, description)
VALUES (gen_random_uuid(), 'venue_staff', 'Venue-scoped staff')
ON CONFLICT DO NOTHING;

CREATE TABLE identity.organizations (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PENDING_VERIFICATION', 'ACTIVE', 'SUSPENDED', 'REJECTED', 'ARCHIVED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX identity_organizations_slug_unique_idx ON identity.organizations (LOWER(slug));

CREATE TABLE identity.organization_members (
    organization_id UUID NOT NULL REFERENCES identity.organizations(id),
    user_id UUID NOT NULL REFERENCES identity.users(id),
    role TEXT NOT NULL CHECK (role IN ('tenant_owner', 'tenant_staff', 'venue_staff')),
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);
CREATE INDEX identity_organization_members_user_idx ON identity.organization_members (user_id, status);

CREATE TABLE identity.venues (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES identity.organizations(id),
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'Asia/Ho_Chi_Minh',
    status TEXT NOT NULL CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'PUBLISHED', 'SUSPENDED', 'ARCHIVED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id, organization_id)
);
CREATE UNIQUE INDEX identity_venues_slug_unique_idx ON identity.venues (LOWER(slug));
CREATE INDEX identity_venues_organization_idx ON identity.venues (organization_id);

-- A venue assignment must belong to the same tenant as the staff membership.
CREATE TABLE identity.venue_staff (
    organization_id UUID NOT NULL,
    venue_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (venue_id, user_id),
    FOREIGN KEY (venue_id, organization_id) REFERENCES identity.venues(id, organization_id),
    FOREIGN KEY (organization_id, user_id) REFERENCES identity.organization_members(organization_id, user_id)
);
CREATE INDEX identity_venue_staff_user_idx ON identity.venue_staff (user_id, status);
