# Identity Service

Identity owns users, credentials, refresh sessions, JWT issuance, and the
`/api/v1/me` query. Its PostgreSQL adapter and SQLC output stay inside this
service boundary.

Identity is an independent Go module (`go.mod`/`go.sum`) and imports only the
infrastructure-only `platform` module plus its own packages. Gateway never
imports Identity code; it calls Identity at runtime through the service
contract.

Its database is configured with `IDENTITY_DATABASE_URL`, its Redis endpoint
with `IDENTITY_REDIS_URL`, and its versioned migrations live in `migrations/`.
Identity never reads another service's database or migration directory.
