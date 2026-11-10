ALTER TABLE catalog.organizations DROP CONSTRAINT organizations_status_check;
ALTER TABLE catalog.organizations ADD CONSTRAINT organizations_status_check
    CHECK (status IN ('PENDING_VERIFICATION', 'ACTIVE', 'SUSPENDED', 'REJECTED', 'ARCHIVED'));
