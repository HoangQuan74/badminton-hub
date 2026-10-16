# Microservice Migration Proposal

## Status

Accepted for Phase 1. Identity/Gateway extraction is implemented locally;
Catalog, Booking, and Payment remain follow-up boundaries and are not deployed
as empty services.

## Problem

Before Phase 1, the backend ran as one Go `api` binary containing the
implemented Identity/Auth slice. It is planned to grow into organization,
venue, booking, payment, realtime, and scheduled-work modules. Phase 1 now
separates the public Gateway from Identity while keeping future business
modules in one source tree until their boundaries are implemented.

The requested direction is to make the backend independently deployable by
business capability without breaking the existing `/api/v1` contract or moving
booking correctness into a distributed lock.

## Goals

- Give each business capability one clear owner for data and state transitions.
- Keep the public HTTP contract stable for the frontend.
- Allow independent deployment and scaling where it has a measurable benefit.
- Preserve PostgreSQL constraints and idempotency for booking and payment flows.
- Introduce service boundaries incrementally with a route-level rollback path.
- Keep local development and the first staging deployment operable by one team.

## Non-goals

- A deployable service for every table or CRUD module.
- Synchronous HTTP calls between every service for one user request.
- Sharing one business database across service boundaries.
- Replacing the current booking transaction with Redis or an application-only
  distributed lock.
- Introducing Kubernetes, service mesh, or multi-region deployment as part of
  the first extraction.

## Current Evidence

- `be/services/gateway/cmd/gateway/main.go` initializes the public Gateway and its Identity reverse
  proxy; it does not open business storage connections.
- `be/services/identity/cmd/identity/main.go` initializes the PostgreSQL pool and Identity router.
- `be/services/gateway/transport/http/router.go` owns only public forwarding and
  readiness; `be/services/identity/transport/http/router.go` owns Identity HTTP
  composition.
- The only implemented business slice is Auth/User. The current endpoints are
  `POST /api/v1/auth/register`, `POST /api/v1/auth/login`,
  `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, and `GET /api/v1/me`.
- `be/services/identity/migrations/000001_create_users_and_sessions.up.sql` owns `users` and
  `auth_sessions`.
- `contracts/openapi/identity.yaml` now captures the public Identity contract;
  a runnable Postman collection is still follow-up work.
- The current local Compose stack has separate `identity` and `api` Gateway
  containers. Identity runs migrations against the dedicated
  `identity-db`/`badminton_identity` database, while Gateway remains stateless
  and has its own Redis endpoint for edge/runtime concerns.
- Gateway and Identity are independent Go modules with service-local
  `go.mod/go.sum`; they share only the infrastructure-only `be/platform`
  module and communicate at runtime through contracts.

## Proposed Topology

```text
                    +------------------+
Web / Admin ------> | Gateway / BFF     | :8080
                    +---------+--------+
                              |
            +-----------------+------------------+
            |                 |                  |
            v                 v                  v
      Identity :8081    Catalog :8082      Booking :8083
      users, sessions   orgs, venues,      holds, bookings,
                        courts, hours,     court blocks,
                        pricing             overlap rules
                              |                  |
                              +--------+---------+
                                       |
                                       v
                                Payment :8084
                                payments, callbacks,
                                refunds, payouts

      Outbox -> Kafka -> Consumer -> notifications / Redis Pub/Sub -> Socket
                                  ^
                                  |
                                Cron
```

The Gateway remains the only public HTTP entry point. It preserves the existing
URLs and forwards a request to the owning service. Services do not expose their
internal ports publicly.

### Service ownership

| Service | Owns | Does not own |
| --- | --- | --- |
| Gateway/BFF | Public routing, request IDs, API error envelope, edge auth checks, rate limits | Business tables or business state transitions |
| Identity | `users`, `auth_sessions`, credentials, token issuance, `/me` | Organization membership decisions for a venue operation |
| Catalog | Organizations, memberships, venues, courts, operating hours, pricing rules | Active court reservations and payment state |
| Booking | Holds, bookings, booking transitions, court blocks, idempotency for booking commands, overlap constraints | User passwords, provider credentials, payment settlement |
| Payment | Payment intents, provider callbacks, payment events, refunds, payouts | Direct writes to booking state |
| Consumer | Outbox publishing, Kafka consumers, notification commands, Kafka-to-Redis fan-out | Authoritative business state |
| Cron | Hold expiration and other scheduled commands | Owning a second copy of booking state |
| Socket | Authenticated connections and local broadcast | Business state and durable event processing |

`court_blocks` belongs with Booking because it occupies the same court timeline
as an online or walk-in booking. This keeps one PostgreSQL overlap constraint
for every active reservation.

## Contracts and Communication

### Public HTTP

The Gateway keeps the current `/api/v1` paths. The first extracted routes are:

```text
POST /api/v1/auth/register  -> Identity
POST /api/v1/auth/login     -> Identity
POST /api/v1/auth/refresh   -> Identity
POST /api/v1/auth/logout    -> Identity
GET  /api/v1/me             -> Identity
```

The response envelope, stable error codes, locale behavior, and status codes
must remain unchanged during extraction. The contract should be captured in an
OpenAPI file and a runnable Postman collection before the Gateway switches
traffic.

### Synchronous service calls

Use gRPC with versioned Protocol Buffer contracts for synchronous business calls
between services, as recorded in
[`ADR-0001`](../adr/0001-grpc-for-internal-service-communication.md). Every call
carries deadlines, `X-Request-ID`, `X-Correlation-ID`, and an authenticated
service identity. Add mTLS or signed service tokens before exposing the network
outside the private application network.

The current Gateway-to-Identity HTTP/JSON proxy is a temporary migration bridge
that preserves the public `/api/v1` contract. Replace it with an Identity gRPC
client/server contract before introducing the next synchronous domain-service
dependency.

Booking may synchronously read Catalog data for operating hours and pricing, but
the final price and the overlap decision remain owned by Booking. A Catalog
failure must fail a hold request clearly; it must never cause the client to
choose a price or bypass the overlap constraint.

### Events

Use versioned events for work that crosses a business boundary:

```text
booking.hold_created.v1
booking.hold_expired.v1
booking.confirmed.v1
booking.cancelled.v1
payment.succeeded.v1
payment.failed.v1
refund.created.v1
notification.requested.v1
```

The event envelope remains the one documented in `docs/05-api-and-events.md`.
Booking and its outbox row commit in one transaction. Payment publishes
`payment.succeeded.v1`; Booking consumes it and performs an idempotent state
transition. Payment never updates the Booking database directly.

## Data and Migration Strategy

### Phase 0: contract and boundary preparation

1. Add `contracts/openapi/`, `contracts/events/`, and versioned
   `contracts/proto/` artifacts for the public Auth endpoints, event envelope,
   and internal gRPC contracts.
2. Add request/correlation IDs to the Gateway-to-Identity request. Authenticated
   service identity remains a hardening follow-up before private network trust
   extends beyond the local deployment.
3. Add service-level health and readiness checks.
4. Keep the current `api` route as the compatibility implementation.

### Phase 1: extract Identity

1. Create `services/identity/cmd/identity` and move Auth composition behind an
   Identity service
   entry point without changing domain behavior.
2. Give Identity its own database and migration directory. A PostgreSQL server
   may host that database alongside future service databases, but the service
   must not share business tables, credentials, or migration history.
3. Run the Gateway and Identity side by side. Route only Auth paths through
   the Gateway to Identity; retain the old API route as a rollback target.
4. Verify register, login, refresh rotation, logout, `/me`, duplicate email,
   inactive user, and invalid refresh-token behavior through the Gateway.
5. After an observation window, remove the duplicate Auth composition from the
   old API process.

The first extraction does not require a data copy. A copy would create two
authoritative user stores and make refresh-token revocation unsafe.

### Phase 2: extract Catalog

Create Catalog only after the owner and membership queries have concrete
handlers and tenant authorization tests. Move organizations, memberships,
venues, courts, operating hours, and pricing together so the service can answer
one coherent venue configuration request.

### Phase 3: extract Booking

Move holds, bookings, walk-ins, blocks, state transitions, and overlap
constraints as one unit. Keep the booking tables and the booking outbox in the
same transaction boundary. This phase must include concurrent hold tests and
idempotency-key reuse tests before traffic is switched.

### Phase 4: extract Payment and workers

Move payment callbacks, provider-event deduplication, refunds, and payouts to
Payment. Keep callback deduplication and payment state in Payment, then notify
Booking through versioned events. Run Consumer and Cron as separate deployables
with their existing idempotent and lock-protected behavior.

### Database boundary

Each service owns a separate PostgreSQL database, credentials, and migration
directory. Identity runs `services/identity/migrations` against
`IDENTITY_DATABASE_URL`; Catalog, Booking, and Payment will add their own
database URL and `<service>/migrations` directory when extracted. One
PostgreSQL server may host several service databases for operational
simplicity, but services must not query another service's tables or run another
service's migrations. Use an API or event instead.

An existing environment that still uses the former shared database needs an
explicit, verified data copy into the new service database before its URL is
switched. The deployment migration job only applies versioned schema files; it
does not copy or delete business data.

Redis configuration is namespaced in the same way (`IDENTITY_REDIS_URL`,
`GATEWAY_REDIS_URL`, `CATALOG_REDIS_URL`, `BOOKING_REDIS_URL`,
`PAYMENT_REDIS_URL`, `NOTIFICATION_REDIS_URL`, `CONSUMER_REDIS_URL`,
`CRON_REDIS_URL`, and `SOCKET_REDIS_URL`). Each value points to a
service-scoped endpoint or logical Redis database; sharing one generic
`DATABASE_URL` or `REDIS_URL` between service processes is not part of the
target topology.

## Security

Identity remains the only token issuer. During Phase 1 the existing JWT secret
may be shared only as a temporary compatibility measure. The target state is an
asymmetric signing key owned by Identity, with Gateway and services validating
the public key through a versioned JWKS endpoint. Internal service calls must
carry an authenticated service identity and must not trust a client-supplied
user or organization ID without verifying the token and scope.

## Rollout and Rollback

Traffic is switched by Gateway route, one route group at a time. Each new route
has a timeout, bounded retry policy, dependency health check, and a metric for
the upstream service and contract error code. Rollback means changing the route
back to the old API implementation; it must not run a destructive migration or
copy data in the reverse direction.

Before switching a route, require:

- contract tests against the new service and the old route;
- service unit and integration tests;
- duplicate and concurrency tests for the affected invariant;
- readiness and shutdown checks;
- a smoke test through the public Gateway URL;
- dashboards and alerts for error rate, latency, dependency failures, and
  event lag.

## Alternatives Considered

### Keep the existing four-process design

This is the lowest-risk option and remains appropriate while only Auth is
implemented. It avoids network calls and distributed ownership, but it does not
provide independently deployable business services.

### Split every planned module immediately

This creates the most operational overhead before the domain code exists. It
would force distributed transactions between Booking, Catalog, and Payment and
would make local debugging and rollback harder.

### Recommended: incremental domain extraction

This preserves the current delivery path while establishing real ownership
boundaries. Identity is the first safe extraction because its data and HTTP
contract are already implemented and its transaction is local to users and
sessions.

## Decision Gate

Approve this proposal only if the team accepts:

1. Gateway is the public entry point.
2. Identity is the first extracted service.
3. Each service has a separate database and migration history. A PostgreSQL
   server may host those databases, but business tables are not shared.
4. Booking remains one correctness boundary, including walk-ins and court
   blocks.
5. The public API contract must be captured before route cutover.

Once approved, the first implementation slice is the Identity service and
Gateway route split. Catalog, Booking, and Payment remain follow-up slices and
are not created as empty microservices.
