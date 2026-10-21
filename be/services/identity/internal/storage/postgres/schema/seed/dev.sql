INSERT INTO identity.roles (id, name, description)
VALUES
    ('00000000-0000-4000-8000-000000000010', 'tenant_owner', 'Tenant organization owner'),
    ('00000000-0000-4000-8000-000000000011', 'tenant_staff', 'Tenant venue staff')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO identity.users (id, email, password_hash, full_name, phone, status)
VALUES (
    '00000000-0000-4000-8000-000000000001',
    'user@badmintonhub.local',
    '$2a$10$T5XMBoQfvSxexUaEtPWrjODidrYn1eBues3d.zCzEDbEsvBAKmxHu',
    'Demo Player',
    '0989250614',
    'active'
)
ON CONFLICT DO NOTHING;

INSERT INTO identity.users (id, email, password_hash, full_name, phone, status, is_admin)
VALUES (
    '00000000-0000-4000-8000-000000000002',
    'admin@badmintonhub.local',
    '$2a$10$T5XMBoQfvSxexUaEtPWrjODidrYn1eBues3d.zCzEDbEsvBAKmxHu',
    'Platform Admin',
    '0900000002',
    'active',
    TRUE
)
ON CONFLICT DO NOTHING;

INSERT INTO identity.users (id, email, password_hash, full_name, phone, status, role_id)
VALUES (
    '00000000-0000-4000-8000-000000000003',
    'owner@badmintonhub.local',
    '$2a$10$T5XMBoQfvSxexUaEtPWrjODidrYn1eBues3d.zCzEDbEsvBAKmxHu',
    'Demo Venue Owner',
    '0900000003',
    'active',
    '00000000-0000-4000-8000-000000000010'
)
ON CONFLICT DO NOTHING;

INSERT INTO identity.users (id, email, password_hash, full_name, phone, status, role_id)
VALUES (
    '00000000-0000-4000-8000-000000000004',
    'staff@badmintonhub.local',
    '$2a$10$T5XMBoQfvSxexUaEtPWrjODidrYn1eBues3d.zCzEDbEsvBAKmxHu',
    'Demo Venue Staff',
    '0900000004',
    'active',
    '00000000-0000-4000-8000-000000000011'
)
ON CONFLICT DO NOTHING;
