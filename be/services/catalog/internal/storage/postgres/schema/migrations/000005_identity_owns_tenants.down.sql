-- Refuse rollback when writes made after cutover cannot be represented by the legacy tables.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM catalog.courts AS c LEFT JOIN catalog.organizations AS o ON o.id = c.organization_id WHERE o.id IS NULL)
      OR EXISTS (SELECT 1 FROM catalog.courts AS c LEFT JOIN catalog.venues AS v ON v.id = c.venue_id WHERE v.id IS NULL)
      OR EXISTS (SELECT 1 FROM catalog.venue_images AS i LEFT JOIN catalog.venues AS v ON v.id = i.venue_id WHERE v.id IS NULL) THEN
        RAISE EXCEPTION 'reconcile Identity tenant and venue data before Catalog rollback';
    END IF;
END $$;
ALTER TABLE catalog.courts ADD CONSTRAINT courts_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES catalog.organizations(id) ON DELETE CASCADE;
ALTER TABLE catalog.courts ADD CONSTRAINT courts_venue_id_fkey FOREIGN KEY (venue_id) REFERENCES catalog.venues(id) ON DELETE CASCADE;
ALTER TABLE catalog.venue_images ADD CONSTRAINT venue_images_venue_id_fkey FOREIGN KEY (venue_id) REFERENCES catalog.venues(id);
DROP TABLE catalog.identity_cutover_state;
DROP TRIGGER freeze_legacy_organizations ON catalog.organizations;
DROP TRIGGER freeze_legacy_members ON catalog.organization_members;
DROP TRIGGER freeze_legacy_venues ON catalog.venues;
DROP FUNCTION catalog.reject_legacy_identity_write();
