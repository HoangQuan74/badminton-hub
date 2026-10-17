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
