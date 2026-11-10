-- Refuse destructive rollback while Identity owns tenant or venue records.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM identity.organizations)
        OR EXISTS (SELECT 1 FROM identity.organization_members)
        OR EXISTS (SELECT 1 FROM identity.venues)
        OR EXISTS (SELECT 1 FROM identity.venue_staff) THEN
        RAISE EXCEPTION 'migrate and verify Identity tenant data before rollback';
    END IF;
END $$;
DROP TABLE identity.venue_staff;
DROP TABLE identity.venues;
DROP TABLE identity.organization_members;
DROP TABLE identity.organizations;
