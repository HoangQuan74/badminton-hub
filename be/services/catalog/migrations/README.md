# Catalog migrations

This directory owns the Catalog service database migration history.

- Use the `CATALOG_DATABASE_URL` when running migrations.
- Keep schema changes and rollback files in this directory.
- Do not read or migrate another service's database.

The service is not deployed yet; this file records the boundary before its first migration is added.
