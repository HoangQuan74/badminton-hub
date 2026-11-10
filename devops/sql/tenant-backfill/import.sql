-- All statements run in one target transaction. Staging rows contain base64
-- text to keep CSV payloads safe even when original values contain newlines.
CREATE TEMP TABLE stage_organizations (id UUID, name64 TEXT, slug64 TEXT, status TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ) ON COMMIT DROP;
CREATE TEMP TABLE stage_members (organization_id UUID, user_id UUID, role TEXT, status TEXT, created_at TIMESTAMPTZ) ON COMMIT DROP;
CREATE TEMP TABLE stage_venues (id UUID, organization_id UUID, name64 TEXT, slug64 TEXT, address64 TEXT, timezone64 TEXT, status TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ) ON COMMIT DROP;
COPY stage_organizations FROM STDIN WITH (FORMAT csv);
__ORGANIZATIONS_CSV__
\.
COPY stage_members FROM STDIN WITH (FORMAT csv);
__MEMBERS_CSV__
\.
COPY stage_venues FROM STDIN WITH (FORMAT csv);
__VENUES_CSV__
\.

INSERT INTO identity.organizations (id, name, slug, status, created_at, updated_at)
SELECT id, convert_from(decode(name64, 'base64'), 'UTF8'), convert_from(decode(slug64, 'base64'), 'UTF8'), status, created_at, updated_at
FROM stage_organizations ON CONFLICT DO NOTHING;
INSERT INTO identity.organization_members (organization_id, user_id, role, status, created_at)
SELECT organization_id, user_id, role, status, created_at FROM stage_members ON CONFLICT DO NOTHING;
INSERT INTO identity.venues (id, organization_id, name, slug, address, timezone, status, created_at, updated_at)
SELECT id, organization_id, convert_from(decode(name64, 'base64'), 'UTF8'), convert_from(decode(slug64, 'base64'), 'UTF8'),
       convert_from(decode(address64, 'base64'), 'UTF8'), convert_from(decode(timezone64, 'base64'), 'UTF8'), status, created_at, updated_at
FROM stage_venues ON CONFLICT DO NOTHING;

-- Existing rows must be identical. Never overwrite Identity-owned records.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM stage_organizations AS s LEFT JOIN identity.organizations AS d ON d.id = s.id
        WHERE d.id IS NULL OR (d.name, d.slug, d.status, d.created_at, d.updated_at)
          IS DISTINCT FROM (convert_from(decode(s.name64, 'base64'), 'UTF8'), convert_from(decode(s.slug64, 'base64'), 'UTF8'), s.status, s.created_at, s.updated_at)
    ) THEN RAISE EXCEPTION 'organization parity mismatch; target transaction rolled back'; END IF;
    IF EXISTS (
        SELECT 1 FROM stage_members AS s LEFT JOIN identity.organization_members AS d ON d.organization_id = s.organization_id AND d.user_id = s.user_id
        WHERE d.user_id IS NULL OR (d.role, d.status, d.created_at) IS DISTINCT FROM (s.role, s.status, s.created_at)
    ) THEN RAISE EXCEPTION 'membership parity mismatch; target transaction rolled back'; END IF;
    IF EXISTS (
        SELECT 1 FROM stage_venues AS s LEFT JOIN identity.venues AS d ON d.id = s.id
        WHERE d.id IS NULL OR (d.organization_id, d.name, d.slug, d.address, d.timezone, d.status, d.created_at, d.updated_at)
          IS DISTINCT FROM (s.organization_id, convert_from(decode(s.name64, 'base64'), 'UTF8'), convert_from(decode(s.slug64, 'base64'), 'UTF8'),
                            convert_from(decode(s.address64, 'base64'), 'UTF8'), convert_from(decode(s.timezone64, 'base64'), 'UTF8'), s.status, s.created_at, s.updated_at)
    ) THEN RAISE EXCEPTION 'venue parity mismatch; target transaction rolled back'; END IF;
END $$;
