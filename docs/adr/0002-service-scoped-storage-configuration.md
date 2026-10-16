---
status: accepted
---

# Use service-scoped database and Redis configuration

Each backend service owns a namespaced storage configuration and credentials.
Identity reads `IDENTITY_DATABASE_URL` and `IDENTITY_REDIS_URL`; future Catalog,
Booking, and Payment services will read their own `CATALOG_*`, `BOOKING_*`, and
`PAYMENT_*` variables. Gateway owns no business database, but its edge/runtime
dependencies use `GATEWAY_*` configuration. New service code must not read a
generic `DATABASE_URL` or `REDIS_URL`, and must not query another service's
tables. The current `api`/Identity Compose compatibility variables exist only
to allow rollback to the pre-extraction image.

## Consequences

- Local, Dev, and Production environment files must provide one connection
  namespace per service.
- Migrations live below the owning service and run only against that service's
  database variable; Identity uses `services/identity/migrations` with
  `IDENTITY_DATABASE_URL`.
- PostgreSQL may remain one server for operational simplicity, but each service
  uses a separate database name, credentials, and migration history. No shared
  business database or cross-service table queries are allowed.
- Redis endpoints or logical databases are isolated per service without
  changing another service's configuration.
- A service that adds a new storage dependency must add its own namespaced
  config, readiness rule, secret, and rollback compatibility before rollout.
