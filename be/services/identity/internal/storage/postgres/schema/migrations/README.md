# Identity migrations

This directory owns the Identity migration history and the `identity`
PostgreSQL schema.

- Use `IDENTITY_DATABASE_URL` when running migrations.
- Keep Identity tables and indexes under the `identity` schema.
- Do not read or migrate another service's schema.

Migration `000002_create_admin_access` adds the Identity-owned admin access
model: `roles`, `permissions`, `role_permissions`, and nullable `users.role_id`
with `users.is_admin`. Admin authorization is evaluated by Identity; gateways do
not query these tables directly.
