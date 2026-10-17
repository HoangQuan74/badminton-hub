# System Design

## Goals

- Correctly prevent double booking.
- Provide one calendar for online and walk-in reservations.
- Keep tenant data isolated.
- Process payment callbacks and events idempotently.
- Remain understandable and deliverable in one week.
- Demonstrate Go concurrency, Kafka, Redis, WebSocket, CI/CD, and observability with justified use cases.

## Non-Goals

- Multi-region availability.
- Real financial settlement.
- Per-tenant infrastructure.

## Architecture

```text
Next.js Web
  |- Public marketplace
  |- Customer account
  |- Owner and staff portal
  `- Platform admin
          |
          v
Three Gateways
  |- User Gateway: public user HTTPS/JSON routes
  |- Admin Gateway: private admin HTTPS/JSON routes
  |- Socket Gateway: public WSS/WebSocket edge
          |
          v
Identity Service (private gRPC)
  |- Users, credentials, refresh sessions
  |- JWT issuance and current-user lookup
          |
          `- Identity PostgreSQL database: identity source of truth

Future domain services
  |- Catalog: organizations, venues, courts, hours, pricing
  |- Booking: holds, bookings, blocks, overlap rules
  `- Payment: intents, callbacks, refunds, payouts
          |- Each stateful service owns a logical PostgreSQL boundary
          |- Each service owns its own Redis endpoint when needed
          `- Each event-producing service owns its own outbox
                 |
                 v
               Kafka
                 |
                 v
Go Consumer
  |- Outbox publishing
  |- Notification processing
  |- Kafka to Redis realtime dispatch
  `- Payment-event processing

Go Cron
  |- Hold expiration
  `- Mock payout processing

Socket Service
  |- Auth handshake
  |- Redis Pub/Sub subscriber
  |- Connection hub
  `- WebSocket broadcast
```

Phase 1 has independently deployable User, Admin, and Socket Gateways plus the
private Identity service under `be`, plus the planned `consumer` and `cron`
processes. Catalog, Booking, and Payment remain modular code until their data
ownership and consistency boundaries are implemented.

Every deployable domain service that owns durable state has a logical PostgreSQL
boundary, service-scoped credentials, migration history, and Redis configuration
when needed. Initial deployments may co-locate service schemas in one
PostgreSQL database for operational simplicity. A service never reads another
service's schema or runs another service's migrations. When independent
scaling is needed, a service can move its schema and migration history to its
own database without changing its domain contract.

## Core Modules

### Identity

- User credentials.
- Access and refresh tokens.
- Global role and organization membership.

### Organization

- Organization lifecycle.
- Owner and staff membership.
- Tenant authorization.

### Venue

- Venue, court, operating hours, amenities, images, and maintenance blocks.

### Availability

- Operating-hour validation.
- Pricing calculation.
- Court-conflict queries.
- Cache invalidation after schedule changes.

### Booking

- Hold creation and expiration.
- State transitions.
- Online and walk-in booking.
- Cancellation and check-in.

### Payment

- Mock payment intent.
- Callback verification.
- Idempotent confirmation.
- Commission, refund, and receivable records.

### Notification and Realtime

- Consumer consumes durable Kafka events idempotently.
- Consumer publishes realtime messages to Redis Pub/Sub.
- Every Socket instance subscribes to Redis Pub/Sub.
- Socket broadcasts only to locally connected authorized clients.
- Clients refetch authoritative API state after reconnect.

## Double-Booking Protection

PostgreSQL is the correctness boundary. Redis must not be the only lock.

Recommended model:

- Persist one time range per hold/booking.
- Treat `PENDING_PAYMENT`, `CONFIRMED`, and `CHECKED_IN` as active reservations.
- Use a PostgreSQL exclusion constraint over `court_id` and a timestamp range.
- Expired holds are transitioned before or during availability checks.

Alternative fixed-slot modeling may use a unique constraint on `(court_id, slot_start)`, but it is less flexible for bookings longer than one slot.

## Idempotency

Sensitive commands require an `Idempotency-Key`:

- Create hold.
- Start payment.
- Cancel booking.
- Create walk-in booking.

Store key, actor, request hash, status, and response. Reusing a key with a different request returns a conflict.

Payment callback deduplication uses the provider event ID and transaction reference. Kafka consumers use unique processed-event records.

## Reliable Events

Booking state and outbox event are committed in one database transaction. The Consumer publishes pending outbox rows to Kafka. Business consumers are at-least-once and idempotent. The realtime dispatcher republishes selected events to Redis Pub/Sub, and every Socket instance receives them for local fan-out.

Consumer replicas scale with Kafka partitions and consumer groups. Cron is a separate deployment, starts with one replica, and protects each scheduled job with a PostgreSQL advisory lock and idempotent state transitions so an accidental second replica cannot duplicate work.

Initial topics:

- `booking.events`
- `payment.events`
- `notification.commands`
- `payout.events`

## Redis Responsibilities

- Search and availability cache.
- Login and booking rate limits.
- WebSocket presence.
- Temporary progress or UI hints.
- Short-lived hold-expiration index as an optimization.

Redis never owns the authoritative booking state.

## MongoDB Responsibilities

MongoDB is connected as an additional document-store capability, but the current MVP assigns it no authoritative business entity. A future reviewed use case may store append-oriented audit documents or raw external-event payloads. Users, bookings, payments, balances, and payout state remain in PostgreSQL. Do not dual-write business state to PostgreSQL and MongoDB without an explicit consistency design.

## Security

- Hash passwords with Argon2id or bcrypt.
- Short-lived access token and rotatable refresh token.
- Enforce organization scope in repository queries.
- Validate payment callback signature.
- Rate-limit authentication and booking endpoints.
- Use server-calculated prices.
- Record audit events for privileged actions.
- Never log passwords, tokens, payment secrets, or complete sensitive payloads.

## Observability

- Structured JSON logs with request ID, actor ID, organization ID, venue ID, booking ID, and event ID when available.
- Prometheus metrics for request rate/error/latency, booking conflicts, hold expiration, payment outcomes, Kafka failures, Consumer health, and Cron runs.
- Grafana dashboards for API, booking funnel, payment, and asynchronous processing.

## Failure Modes

| Failure | Required behavior |
|---|---|
| Two customers hold the same slot | Database accepts only one active reservation |
| Payment callback is duplicated | Booking confirms once; duplicate returns success safely |
| Payment succeeds after hold expiry | Do not silently confirm; record exception for reconciliation |
| Kafka is unavailable | Booking transaction succeeds with outbox pending for retry |
| Redis is unavailable | Core booking remains correct; cache and rate-limit behavior degrade safely |
| Consumer restarts | Pending outbox and Kafka processing resume idempotently |
| Cron restarts | Due jobs resume from database truth without duplicate state transitions |
| WebSocket disconnects | Client refetches authoritative snapshot after reconnect |
