-- Refuse rollback before changing the constraint if archived records remain.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM catalog.venues WHERE status = 'ARCHIVED') THEN
        RAISE EXCEPTION 'restore archived venues before rolling back';
    END IF;
END $$;

-- Rollback is intentionally blocked while archived records exist; restore them explicitly first.
ALTER TABLE catalog.venues DROP CONSTRAINT venues_status_check;
ALTER TABLE catalog.venues ADD CONSTRAINT venues_status_check
    CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'PUBLISHED', 'SUSPENDED'));
