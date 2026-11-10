-- Refuse rollback before changing the constraint if archived records remain.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM catalog.organizations WHERE status = 'ARCHIVED') THEN
        RAISE EXCEPTION 'restore archived organizations before rolling back';
    END IF;
END $$;

-- Restore archived records explicitly before rolling back this constraint.
ALTER TABLE catalog.organizations DROP CONSTRAINT organizations_status_check;
ALTER TABLE catalog.organizations ADD CONSTRAINT organizations_status_check
    CHECK (status IN ('PENDING_VERIFICATION', 'ACTIVE', 'SUSPENDED', 'REJECTED'));
