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
- Sharing business tables or schemas across service boundaries.
- Replacing the current booking transaction with Redis or an application-only
  distributed lock.
- Introducing Kubernetes, service mesh, or multi-region deployment as part of
  the first extraction.

## Current Evidence

- `be/gateways/user/cmd/user-gateway/main.go` initializes the public User
  Gateway; it does not open business storage or own Identity state
  connections.
- `be/services/identity/cmd/identity/main.go` initializes the PostgreSQL pool and Identity router.
- User and Admin Gateway routes call Identity over gRPC. Socket Gateway now
  dials a gRPC target and fails closed until the versioned Socket stream
  contract is registered. Each gateway folder owns only its route set and
  transport configuration.
  Identity owns the private gRPC server composition and no longer starts an HTTP
  compatibility listener.
- The only implemented business slice is Auth/User. The current endpoints are
  `POST /api/v1/auth/register`, `POST /api/v1/auth/login`,
  `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, and `GET /api/v1/me`.
- `be/services/identity/internal/storage/postgres/schema/migrations/000001_create_users_and_sessions.up.sql` owns `users` and
  `auth_sessions`.
- `contracts/openapi/user-gateway.yaml` captures the public User Gateway
  contract;
  a runnable Postman collection is still follow-up work.
- The current local Compose stack has separate User/Admin/Socket Gateway and
  Identity containers. Identity runs migrations against the Identity storage
  boundary, while gateways remain stateless and use only their edge/runtime
  dependencies.
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

User Gateway and Admin Gateway are the public HTTPS/JSON entry points, and
Socket Gateway is the public WSS entry point. The frontend never calls a
service directly. Gateways forward business requests to the owning service over
versioned gRPC contracts. Services do not expose their internal ports publicly.

### Service ownership

| Service | Owns | Does not own |
| --- | --- | --- |
| User/Admin Gateway | Public HTTPS/JSON routing, request IDs, API error envelope, edge auth checks, rate limits | Business tables or business state transitions |
| Socket Gateway | WSS protocol, connection lifecycle, request IDs, edge auth | Business tables or durable event state |
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
POST /api/v1/auth/register  -> Gateway HTTP -> Identity gRPC Register
POST /api/v1/auth/login     -> Gateway HTTP -> Identity gRPC Login
POST /api/v1/auth/refresh   -> Gateway HTTP -> Identity gRPC Refresh
POST /api/v1/auth/logout    -> Gateway HTTP -> Identity gRPC Logout
GET  /api/v1/me             -> Gateway HTTP -> Identity gRPC GetCurrentUser
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

The current Gateway-to-Identity HTTP/JSON proxy is temporary migration code. It
must not be treated as an internal service API or extended with new business
routes. Replace it with the versioned Identity gRPC client/server contract
before adding the next synchronous domain-service dependency.

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
2. Add request/correlation IDs to the Gateway-to-Identity gRPC metadata. Authenticated
   service identity remains a hardening follow-up before private network trust
   extends beyond the local deployment.
3. Add service-level health and readiness checks.
4. Keep the current `api` route as the compatibility implementation.

### Phase 1: extract Identity

1. Create `services/identity/cmd/identity` and move Auth composition behind an
   Identity service
   entry point without changing domain behavior.
2. Give Identity the logical `identity` storage boundary and its migration
   directory. The initial deployment may co-locate the Identity schema with
   future service schemas in one PostgreSQL database, but the service must not
   share business tables, credentials, or migration history. The connection can
   move to a dedicated Identity database later without changing the service
   contract.
3. Run the Gateway and Identity gRPC server side by side. Route only Auth paths
   through the new gRPC client; retain the HTTP bridge as a rollback target.
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

Each stateful service owns a logical PostgreSQL boundary named after the
service, credentials/configuration, and migration history. Identity runs
`services/identity/internal/storage/postgres/schema/migrations` against
`IDENTITY_DATABASE_URL`; Catalog, Booking, and Payment will add their own
storage connection and migration directory when extracted. Initial deployment
may co-locate those schemas in one PostgreSQL database, but services must not
query another service's schema or run another service's migrations. Use gRPC or
events instead. Moving a service to a dedicated database is an operational
change with no domain contract change.

An existing environment that still uses the former shared tables needs an
explicit, verified migration into the owning service schema before the schema
boundary is enforced. The deployment migration job only applies versioned
schema files; it does not copy or delete business data.

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
3. Each stateful service has a logical storage boundary named after the service
   and its own migration history. Initial deployment may co-locate schemas in
   one PostgreSQL database, but business tables and schema access are not
   shared. Each service can move to a dedicated database when scaling requires
   it without changing its gRPC or event contracts.
4. Booking remains one correctness boundary, including walk-ins and court
   blocks.
5. The public API contract must be captured before route cutover.

Once approved, the first implementation slice is the Identity service and
Gateway route split. Catalog, Booking, and Payment remain follow-up slices and
are not created as empty microservices.


## Current blockers before deleting the bridge

The migration is not complete because two public gateway surfaces still have
service implementation gaps:

- Admin publishes Auth, Staff, Role, Permission, and admin password routes.
  Admin Gateway now calls Identity through gRPC; Identity serves Staff CRUD,
  while Admin Auth, Role, Permission, and password operations still return
  `FEATURE_NOT_IMPLEMENTED` until their Identity domain handlers exist.
- Socket publishes `/ws` and bridges it to the versioned Socket gRPC stream;
  the service owns JWT handshake validation and Redis fan-out.

Identity HTTP compatibility transport and the shared HTTP gateway proxy have
been removed. The Socket gRPC stream, Redis fan-out service and WSS bridge are implemented;
remaining evidence is runtime Compose verification.

## Legacy removal gates

The HTTP migration bridge is removed only after each gate has evidence:

1. `contracts/proto/identity/v1/auth.proto` has generated Go client/server
   code under `contracts/gen/identity/v1` and generation is reproducible in CI.
2. Identity serves the Auth gRPC methods with deadlines, authenticated service
   metadata, stable status mapping, and gRPC health status.
3. User Gateway translates its public HTTPS/JSON Auth contract to the Identity
   gRPC client. The source-level route cutover is complete for register, login,
   refresh, logout, and `/me`; focused endpoint and dependency checks remain
   required before deleting the compatibility HTTP transport.
4. Admin Auth/Staff/Role/Permission domain services, gRPC methods, and
   authorization checks exist before Admin Gateway routes can return business
   results instead of `FEATURE_NOT_IMPLEMENTED`.
5. Compose and release manifests use gRPC targets for User, Admin, and Socket
   Gateways. Socket must register its versioned stream service before `/ws` is
   enabled.
