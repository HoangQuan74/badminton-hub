# AI Context Index

This file is the entry point for AI-assisted development. Read the documents in the order below before proposing architecture or writing code.

## Required Reading Order

1. `../../README.md` - documentation map.
2. `../analysis/00-business-context.md` - product and revenue model.
3. `../analysis/03-customer-booking.md` - highest-risk business flow.
4. `../design/01-high-level-design.md` - component boundaries.
5. `../design/05-folder-structure.md` - repository boundaries and ownership.
6. `../design/04-booking-state-machine.md` - allowed state transitions.
7. `../04-data-model.md` - entity and constraint reference.
8. `../05-api-and-events.md` - API and event contracts.
9. `../06-testing-and-devops.md` - quality and operational requirements.
10. `../08-backend-engineering-practices.md` - backend implementation and review standards.
11. `../07-delivery-plan.md` - implementation order.

## Locked Decisions

- Product name: BadmintonHub.
- Model: public B2B2C marketplace plus multi-tenant owner portal.
- One owner organization may manage multiple venues.
- One venue contains multiple courts.
- Online and walk-in bookings share one availability source.
- PostgreSQL is the source of truth and prevents overlaps.
- Redis is an optimization, never the booking correctness boundary.
- MongoDB is provisioned as an infrastructure client but currently owns no business data. PostgreSQL remains authoritative.
- Phase 1 uses a public `api` Gateway/BFF and a private `identity` service under `be`; `socket`, `consumer`, and `cron` remain separate process boundaries. Catalog, Booking, and Payment are future service extractions governed by `docs/design/06-microservice-migration.md`.
- Public Gateway routes use HTTP/JSON. Synchronous business communication between backend services uses versioned gRPC/Protobuf contracts; asynchronous cross-service workflows use Kafka events. The Phase 1 Gateway-to-Identity HTTP proxy is a temporary migration bridge.
- Service configuration is namespaced by owner: Identity uses `IDENTITY_*`, Gateway uses `GATEWAY_*`, future Catalog/Booking/Payment services use their own prefixes, and process boundaries use prefixes such as `CONSUMER_*`, `CRON_*`, and `SOCKET_*`. Services do not share generic database or Redis configuration variables.
- Each service owns a separate database and migration history. Identity uses `services/identity/migrations` with `IDENTITY_DATABASE_URL`; future services add their own migration directory and database URL. A PostgreSQL server may host multiple service databases, but business tables are never shared.
- Consumer replicas scale with Kafka partitions. Cron is deployed separately, starts with one replica, and uses distributed locking plus idempotent jobs.
- Booking and outbox event are committed atomically.
- Kafka consumers must be idempotent.
- Consumer republishes selected durable events to Redis Pub/Sub for realtime fan-out.
- Auth/User is the first implemented vertical slice inside Identity: Gin handlers call the auth domain service, which depends on a repository interface implemented in `services/identity/adapters/postgres`.
- Gateway wiring follows `services/gateway/cmd/gateway -> Gateway router -> Identity HTTP contract`; Identity wiring follows `services/identity/cmd/identity -> Identity router -> feature handler -> domain service -> repository`.
- HTTP request and response contracts live in `services/identity/transport/http/models/request` and `services/identity/transport/http/models/response`, not inside handlers.
- Error messages support English and Vietnamese through embedded go-i18n catalogs. `?lang` takes precedence over `Accept-Language`; English is the fallback.
- Domain errors expose stable message IDs. Handlers return errors; the API endpoint wrapper chooses HTTP status/API code, localizes messages, and logs unexpected failures.
- Use `HandleJSON[T]` for JSON endpoints and `Handle` for endpoints without a JSON body. Do not repeat binding or domain-error switches inside handlers.
- Shared helpers must remain narrow: password hashing is in `pkg/helper`; email normalization is in `pkg/utils`.
- Registration persists the user and refresh session atomically. Passwords use bcrypt; opaque refresh tokens are stored only as SHA-256 hashes and rotate on refresh.
- PostgreSQL access uses handwritten sqlc queries in `services/identity/storage/query` and generated pgx/v5 code in `services/identity/storage/sqlc`; generated files are never edited manually.
- Gateway and Identity are independent Go modules with their own `go.mod` and
  `go.sum`. They may import the infrastructure-only `platform` module, but
  never each other's packages or module paths.
- The root `be` directory is orchestration/documentation only; it has no Go
  module. Cross-service code reuse happens through gRPC/Protobuf or Kafka
  contracts, not Go imports.
- Schema changes use `golang-migrate` versioned up/down files and run as an explicit deployment step, not at API startup.
- Local development uses the all-in-one Compose stack and development seed. Dev and Production never run the seed job.
- Release images are immutable and identified by Git SHA. `develop` auto-deploys Dev; `v*` tags require approval before Production deployment.
- The initial server topology is one Web VM for stateless API/Web workloads and one DB VM for PostgreSQL, Redis, and MongoDB on a private network.
- Application rollback changes image tags only and never runs automatic migration down. Schema migrations must remain backward-compatible.
- K3s may run stateless API/Web workloads on the Web VM for learning, but the two-VM topology is not Kubernetes HA and the DB VM must not be reused as a general application worker.
- Gateway shutdown handles `SIGINT` and `SIGTERM`, drains HTTP requests within a configured timeout, and its readiness follows Identity.
- Identity startup pings PostgreSQL. Its readiness requires PostgreSQL; storage clients for other process boundaries are not initialized by Identity.
- Socket instances subscribe to Redis Pub/Sub; they do not share a Kafka consumer group for client broadcast.
- WebSocket reconnect always refetches authoritative state from the API.
- Prices are calculated and snapshotted by the server.
- All timestamps are stored in UTC; venues have an IANA timezone.
- MVP payment and payout are simulated.

## Implementation Priorities

1. Tenant authorization.
2. Availability and overlap protection.
3. Booking state machine and idempotency.
4. Payment callback and hold expiration.
5. Tests for concurrent and duplicate requests.
6. CI/CD and observability.
7. UI polish and advanced features.

## AI Working Rules

- Do not invent business rules outside the documented scope.
- Do not create additional deployable services without a reviewed ownership, contract, migration, and rollback decision.
- Do not use Redis locks as the only double-booking protection.
- Do not accept final prices from the frontend.
- Do not skip tenant filters in owner or staff repository methods.
- Update the relevant document when a contract or locked decision changes.
- Keep code changes small, testable, and aligned with the delivery plan.
- Apply the backend review checklist before completing a backend change.

## Open Decisions

- Exact Go module path.
- SQL access choice: SQLC or pgx repository layer.
- Authentication library and password-hashing algorithm.
- Kafka implementation for local development.
- AWS deployment target: EC2 first, ECS later if needed.
