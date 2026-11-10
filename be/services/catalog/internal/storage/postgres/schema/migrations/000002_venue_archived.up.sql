ALTER TABLE catalog.venues DROP CONSTRAINT venues_status_check;
ALTER TABLE catalog.venues ADD CONSTRAINT venues_status_check
    CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'PUBLISHED', 'SUSPENDED', 'ARCHIVED'));
