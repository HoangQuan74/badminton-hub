CREATE TABLE identity.tenant_registrations (
    organization_id UUID PRIMARY KEY REFERENCES identity.organizations(id) ON DELETE CASCADE,
    contact_name TEXT NOT NULL,
    contact_email TEXT NOT NULL,
    contact_phone TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO identity.tenant_registrations (organization_id, contact_name, contact_email, contact_phone)
SELECT o.id, u.full_name, u.email, COALESCE(u.phone, '')
FROM identity.organizations o
JOIN identity.organization_members m ON m.organization_id = o.id AND m.role = 'tenant_owner'
JOIN identity.users u ON u.id = m.user_id
ON CONFLICT DO NOTHING;

CREATE TABLE identity.tenant_registration_reviews (
    organization_id UUID PRIMARY KEY REFERENCES identity.tenant_registrations(organization_id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES identity.users(id),
    decision TEXT NOT NULL CHECK (decision IN ('ACTIVE', 'REJECTED')),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
