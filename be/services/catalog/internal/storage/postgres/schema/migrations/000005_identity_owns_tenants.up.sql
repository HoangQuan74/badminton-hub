-- Identity owns organizations and venues. Keep legacy tables for an explicit
-- backfill and rollback window; do not add cross-service database references.
ALTER TABLE catalog.courts DROP CONSTRAINT courts_organization_id_fkey;
ALTER TABLE catalog.courts DROP CONSTRAINT courts_venue_id_fkey;
ALTER TABLE catalog.venue_images DROP CONSTRAINT venue_images_venue_id_fkey;

-- A nonempty legacy directory requires an explicit, verified transfer.
CREATE TABLE catalog.identity_cutover_state (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    ready BOOLEAN NOT NULL,
    verified_at TIMESTAMPTZ
);
INSERT INTO catalog.identity_cutover_state (singleton, ready)
SELECT TRUE, NOT EXISTS (SELECT 1 FROM catalog.organizations)
             AND NOT EXISTS (SELECT 1 FROM catalog.organization_members)
             AND NOT EXISTS (SELECT 1 FROM catalog.venues);

-- Prevent old binaries from silently writing a second tenant authority after
-- the source snapshot or before the Identity cutover is verified.
CREATE FUNCTION catalog.reject_legacy_identity_write() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'legacy tenant and venue writes are disabled for Identity cutover';
END $$;
CREATE TRIGGER freeze_legacy_organizations BEFORE INSERT OR UPDATE OR DELETE OR TRUNCATE ON catalog.organizations FOR EACH STATEMENT EXECUTE FUNCTION catalog.reject_legacy_identity_write();
CREATE TRIGGER freeze_legacy_members BEFORE INSERT OR UPDATE OR DELETE OR TRUNCATE ON catalog.organization_members FOR EACH STATEMENT EXECUTE FUNCTION catalog.reject_legacy_identity_write();
CREATE TRIGGER freeze_legacy_venues BEFORE INSERT OR UPDATE OR DELETE OR TRUNCATE ON catalog.venues FOR EACH STATEMENT EXECUTE FUNCTION catalog.reject_legacy_identity_write();
