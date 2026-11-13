-- System roles must exist without development seeds. Keep existing role IDs.
INSERT INTO identity.roles (id, name, description)
VALUES
    ('00000000-0000-4000-8000-000000000010', 'tenant_owner', 'Tenant organization owner'),
    ('00000000-0000-4000-8000-000000000011', 'tenant_staff', 'Tenant-wide staff')
ON CONFLICT DO NOTHING;

-- The operation scope is implicit: this table stores only owner enrollment.
CREATE TABLE identity.onboard_requests (
    actor_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    idempotency_key UUID NOT NULL,
    request_hash BYTEA NOT NULL CHECK (octet_length(request_hash) = 32),
    response JSONB NOT NULL CHECK (jsonb_typeof(response) = 'object'),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (actor_id, idempotency_key)
);
CREATE INDEX onboard_requests_expiry_idx ON identity.onboard_requests (expires_at);
