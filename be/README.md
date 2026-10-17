# BadmintonHub Backend

## Services

The backend has Identity plus three independent gateway processes. Each has an
independent Go module:

- `services/identity`: owns users, credentials, refresh sessions, and private
  gRPC Auth methods behind the public User Gateway.
  Its module is `services/identity/go.mod`.
- `gateways/user`: public user-facing gateway; its entrypoint is
  `cmd/user-gateway`.
- `gateways/admin`: private admin gateway; its entrypoint is
  `cmd/admin-gateway`.
- `gateways/socket`: WebSocket edge gateway; its entrypoint is
  `cmd/socket-gateway`.

Identity owns the `identity` PostgreSQL schema. Gateway processes do not connect
to business storage; each gateway readiness follows its configured upstream.
Each process reads only its own namespaced configuration.

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
GET  /metrics
```

Each HTTP gateway exposes Prometheus text metrics at `/metrics`, including
request count, duration histogram, and in-flight requests. Labels are limited
to `service`, HTTP method, registered route, and status code.

Error messages support English and Vietnamese. Select a locale with `Accept-Language: vi` or `?lang=vi`; English is the fallback. API error `code` values remain stable across languages.

Registration creates the user and refresh session in one transaction. Access tokens are signed JWTs. Refresh tokens are random opaque values; only their SHA-256 hashes are stored. Refresh rotates the session atomically, and logout revokes it.

## Run locally

1. From the repository root, start the local dependencies:

```bash
docker compose -f devops/docker/compose.yml up -d postgres redis mongodb
```

2. Apply migrations with `IDENTITY_DATABASE_URL=... make migrate-up`.
3. Copy `.env.example` values into your shell and replace `IDENTITY_JWT_SECRET`.
4. Start Identity:

```bash
(cd services/identity && go run ./cmd/identity)
```

5. Start the Gateway in another shell:

```bash
(cd gateways/user && \
  USER_GATEWAY_HTTP_ADDR=:8081 USER_GATEWAY_IDENTITY_GRPC_TARGET=127.0.0.1:9091 \
  go run ./cmd/user-gateway)
```

Identity listens only on private gRPC `:9091` for internal service calls. The User Gateway listens on `:8081` and calls Identity
over the versioned gRPC contract. Send `SIGINT` or `SIGTERM` to stop accepting new requests and
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

Set `IDENTITY_DATABASE_URL` before running migration commands. Do not edit files in `services/identity/internal/storage/postgres/generated/sqlc`; change `services/identity/internal/storage/postgres/query/*.sql` or `services/identity/internal/storage/postgres/schema/migrations` and regenerate them instead. Migrations are an explicit deployment step and are not executed automatically by each API instance.

## Data ownership

- PostgreSQL is the source of truth for each service's owned state. Each
  stateful service owns a logical database boundary, schema, credentials, and
  migration directory. Identity currently uses the database selected by
  `IDENTITY_DATABASE_URL`; initial environments may co-locate service schemas,
  but no service connects to or queries another service's schema. Future
  services can move to separate databases without changing service contracts.
- PostgreSQL, MongoDB, and Redis client lifecycle helpers live together under
  `be/platform/database`; connection URLs remain service-scoped (`IDENTITY_*`,
  `GATEWAY_*`, and future service prefixes) for caching, rate limiting,
  presence, and realtime fan-out. Redis must not be the only correctness
  boundary.
- MongoDB is connected but has no business collection yet. A suitable future use is append-oriented audit or external-event payload storage. Do not duplicate authoritative booking or payment state into MongoDB without a reviewed design.

## Internationalization

- Domain errors expose stable translation keys, not user-facing sentences.
- Gin i18n middleware builds a request-scoped localizer from `lang` and `Accept-Language`.
- `HandleJSON[T]` handles JSON decoding and validation, invokes typed handlers that return `error`, then one endpoint wrapper maps domain errors to HTTP status, stable API code, localized message, and centralized logging.
- Add the same message ID to both English and Vietnamese catalogs before using it in a response.
