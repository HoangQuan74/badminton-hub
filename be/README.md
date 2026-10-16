# BadmintonHub Backend

## Services

The backend now has two deployable HTTP processes for the first extraction,
each with an independent Go module:

- `services/identity`: owns users, credentials, refresh sessions, and `/api/v1/me`.
  Its module is `services/identity/go.mod`.
- `services/gateway`: public Gateway/BFF that preserves `/api/v1` and proxies Auth routes
  to Identity.
  Its module is `services/gateway/go.mod`.

Identity owns PostgreSQL access. The Gateway does not connect to business
storage; its readiness follows Identity readiness. Each process reads only its
own namespaced configuration (`IDENTITY_*` or `GATEWAY_*`).

## Auth/User slice

Implemented endpoints:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/me
GET  /health
GET  /ready
```

Error messages support English and Vietnamese. Select a locale with `Accept-Language: vi` or `?lang=vi`; English is the fallback. API error `code` values remain stable across languages.

Registration creates the user and refresh session in one transaction. Access tokens are signed JWTs. Refresh tokens are random opaque values; only their SHA-256 hashes are stored. Refresh rotates the session atomically, and logout revokes it.

## Run locally

1. From the repository root, start the local dependencies:

```bash
docker compose -f devops/docker/compose.yml up -d identity-db identity-redis gateway-redis mongodb
```

2. Apply migrations with `IDENTITY_DATABASE_URL=... make migrate-up`.
3. Copy `.env.example` values into your shell and replace `IDENTITY_JWT_SECRET`.
4. Start Identity:

```bash
(cd services/identity && go run ./cmd/identity)
```

5. Start the Gateway in another shell:

```bash
(cd services/gateway && \
  GATEWAY_HTTP_ADDR=:8081 GATEWAY_IDENTITY_URL=http://127.0.0.1:8080 \
  go run ./cmd/gateway)
```

Identity listens on `:8080` and the Gateway listens on `:8081` in this direct
run example. Send `SIGINT` or `SIGTERM` to stop accepting new requests and
drain in-flight requests within `SHUTDOWN_TIMEOUT`.

Identity initializes PostgreSQL and its `/ready` endpoint succeeds only when
PostgreSQL responds. Gateway `/ready` succeeds only when Identity is ready.
Both processes drain in-flight HTTP requests within `SHUTDOWN_TIMEOUT`.

## Verify

```bash
make test
```

## Database workflow

Database schema changes use `golang-migrate`. Application queries use `sqlc` with `pgx/v5`.
The root Makefile orchestrates the independent service modules; it is not a Go
module itself.

```bash
# Create a new up/down migration pair
make migrate-create name=add_example

# Apply or roll back migrations
make migrate-up
make migrate-down
make migrate-version

# Regenerate typed Go code after changing SQL
make generate
make sqlc-check
```

Set `IDENTITY_DATABASE_URL` before running migration commands. Do not edit files in `services/identity/storage/sqlc`; change `services/identity/storage/query/*.sql` or migrations and regenerate them instead. Migrations are an explicit deployment step and are not executed automatically by each API instance.

## Data ownership

- PostgreSQL is the source of truth for each service's owned state. Identity uses
  `badminton_identity` locally and `IDENTITY_DATABASE_URL`; future services get
  their own database URL and migration directory. No service queries another
  service's database.
- Redis is configured per service (`IDENTITY_REDIS_URL`, `GATEWAY_REDIS_URL`, and
  future service prefixes) for caching, rate limiting, presence, and realtime
  fan-out. It must not be the only correctness boundary.
- MongoDB is connected but has no business collection yet. A suitable future use is append-oriented audit or external-event payload storage. Do not duplicate authoritative booking or payment state into MongoDB without a reviewed design.

## Internationalization

- Domain errors expose stable translation keys, not user-facing sentences.
- Gin i18n middleware builds a request-scoped localizer from `lang` and `Accept-Language`.
- `HandleJSON[T]` handles JSON decoding and validation, invokes typed handlers that return `error`, then one endpoint wrapper maps domain errors to HTTP status, stable API code, localized message, and centralized logging.
- Translation catalogs are embedded from `services/identity/transport/http/i18n/locales` into the API binary.
- Add the same message ID to both English and Vietnamese catalogs before using it in a response.
