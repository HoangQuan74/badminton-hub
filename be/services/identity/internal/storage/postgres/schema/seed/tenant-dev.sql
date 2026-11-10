-- Identity is the authority for the local tenant, venues, and staff scope.
INSERT INTO identity.organizations (id, name, slug, status)
VALUES ('10000000-0000-4000-8000-000000000001', 'Smash Center', 'smash-center', 'ACTIVE')
ON CONFLICT (id) DO NOTHING;
INSERT INTO identity.organization_members (organization_id, user_id, role, status)
VALUES
    ('10000000-0000-4000-8000-000000000001', '00000000-0000-4000-8000-000000000003', 'tenant_owner', 'ACTIVE'),
    ('10000000-0000-4000-8000-000000000001', '00000000-0000-4000-8000-000000000004', 'tenant_staff', 'ACTIVE'),
    ('10000000-0000-4000-8000-000000000001', '00000000-0000-4000-8000-000000000005', 'venue_staff', 'ACTIVE')
ON CONFLICT (organization_id, user_id) DO NOTHING;
INSERT INTO identity.venues (id, organization_id, name, slug, address, timezone, status)
VALUES
    ('20000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001', 'Smash Center', 'smash-center', 'Bình Thạnh, TP.HCM', 'Asia/Ho_Chi_Minh', 'PUBLISHED'),
    ('20000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001', 'Smash Center Quận 7', 'smash-center-quan-7', 'Quận 7, TP.HCM', 'Asia/Ho_Chi_Minh', 'PUBLISHED'),
    ('20000000-0000-4000-8000-000000000003', '10000000-0000-4000-8000-000000000001', 'Smash Center Thủ Đức', 'smash-center-thu-duc', 'Thủ Đức, TP.HCM', 'Asia/Ho_Chi_Minh', 'DRAFT')
ON CONFLICT (id) DO NOTHING;
INSERT INTO identity.venue_staff (organization_id, venue_id, user_id, status)
VALUES ('10000000-0000-4000-8000-000000000001', '20000000-0000-4000-8000-000000000001', '00000000-0000-4000-8000-000000000005', 'ACTIVE')
ON CONFLICT (venue_id, user_id) DO NOTHING;
