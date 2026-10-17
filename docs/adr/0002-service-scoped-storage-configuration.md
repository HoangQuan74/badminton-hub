---
status: accepted
---

# Use service-owned storage boundaries and service-scoped configuration

Each backend service owns a namespaced storage configuration and credentials.
Identity reads `IDENTITY_DATABASE_URL` and `IDENTITY_REDIS_URL` and owns the
Identity logical PostgreSQL boundary; future Catalog,
Booking, and Payment services will read their own `CATALOG_*`, `BOOKING_*`, and
`PAYMENT_*` variables. Gateway owns no business database, but its edge/runtime
dependencies use `GATEWAY_*` configuration. New service code must not read a
generic `DATABASE_URL` or `REDIS_URL`, and must not query another service's
database or tables. The current `api`/Identity Compose compatibility variables
exist only to allow rollback to the pre-extraction image.

## Consequences

- Local, Dev, and Production environment files must provide one connection
  namespace per service.
- Migrations live below the owning service and target only that service's
  database; Identity uses
  `services/identity/internal/storage/postgres/schema/migrations` with
  `IDENTITY_DATABASE_URL`.
- Each stateful service owns its own logical PostgreSQL boundary, credentials,
  schema, and migration history. Initial environments may co-locate service
  schemas in one PostgreSQL database, but a service never connects to or
  queries another service's schema. Moving a service to its own database is an
  infrastructure change with no domain contract change.
- Redis endpoints or logical databases are isolated per service without
  changing another service's configuration.
- A service that adds a new storage dependency must add its own namespaced
  config, readiness rule, secret, and rollback compatibility before rollout.
