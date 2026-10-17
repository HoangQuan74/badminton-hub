# Testing and DevOps

## Test Strategy

### Unit tests

- Pricing rules and price snapshots.
- Booking-state transitions.
- Cancellation and refund calculations.
- Role and tenant authorization.
- Hold-expiration decisions.
- Payment-callback idempotency.

Use table-driven Go tests with Testify or GoConvey where it improves readability.

### Integration tests

- PostgreSQL overlap constraint.
- Concurrent hold attempts for one court and time range.
- Transactional booking plus outbox insertion.
- Idempotency-key reuse.
- Kafka consumer deduplication.
- Redis cache invalidation and fallback behavior.
- Kafka-to-Redis realtime dispatch for selected events.
- Redis Pub/Sub fan-out reaches multiple Socket instances.

Prefer real PostgreSQL, Redis, and Kafka containers over mocked infrastructure for these tests.

### End-to-end tests

Required scenarios:

1. Owner creates and publishes a venue after admin approval.
2. Customer searches, holds, pays, and receives a confirmed booking.
3. Two concurrent customers attempt the same slot; one receives `BOOKING_SLOT_UNAVAILABLE`.
4. Duplicated payment callback confirms once.
5. Unpaid hold expires and becomes available again.
6. Staff walk-in booking blocks online availability.

## Local Environment

Docker Compose services:

- `web`
- `user-gateway`
- `admin-gateway`
- `socket-gateway`
- `identity`
- `socket`
- `consumer`
- `cron`
- `postgres`
- `redis`
- `kafka`
- `prometheus`
- `grafana`
- optional `mailpit`

The local stack must start with one documented command and include deterministic seed data.

## Jenkins Pipeline

The shared `be/platform` module has a module-wide coverage gate of 100%. The
gate instruments all platform packages with `-coverpkg=./...`; tests that only
cover the `be/platform/test` package do not inflate the result. Run it locally
with:

```bash
cd be
make coverage-platform
```

```text
Checkout
-> dependency restore
-> format and lint
-> unit tests
-> integration tests
-> frontend type-check and build
-> API, Socket, Consumer, and Cron builds
-> Docker image build
-> image vulnerability scan
-> push images
-> deploy staging
-> smoke tests
```

Production deployment is not required for the MVP. A reproducible staging deployment is required.

## Deployment

Initial target:

- AWS EC2.
- Docker Compose.
- Nginx or Caddy reverse proxy.
- Managed DNS.
- HTTPS.
- GitHub Container Registry or Amazon ECR.

Future targets may include ECS or Kubernetes, but the first version should remain operationally simple.

## Required Prometheus Metrics

- `http_requests_total`
- `http_request_duration_seconds`
- `booking_holds_total`
- `booking_conflicts_total`
- `booking_hold_expired_total`
- `booking_confirmed_total`
- `payment_callbacks_total`
- `payment_callback_duplicates_total`
- `outbox_pending_total`
- `kafka_consumer_errors_total`
- `cron_job_runs_total`
- `cron_job_lock_skipped_total`
- `websocket_connections`
- `websocket_messages_total`
- `websocket_slow_clients_total`
- `redis_pubsub_errors_total`

## Grafana Dashboards

### API

- Request rate.
- Error rate.
- p50, p95, and p99 latency.
- Slowest endpoints.

### Booking

- Holds, confirmations, expirations, and conflicts.
- Hold-to-confirm conversion.
- Online versus walk-in bookings.

### Asynchronous processing

- Pending outbox events.
- Publish attempts and failures.
- Kafka consumer errors and lag where available.

### Socket

- Active connections and rooms.
- Message rate and Redis Pub/Sub errors.
- Slow-client disconnects and dropped non-critical events.

### Cron

- Job duration, success, and failure counts.
- Distributed-lock skips and overdue runs.

## Logging

Use structured JSON logs. Include identifiers only when relevant:

- `request_id`
- `correlation_id`
- `user_id`
- `organization_id`
- `venue_id`
- `court_id`
- `booking_id`
- `event_id`

Do not log passwords, tokens, secrets, or complete payment payloads.

## Operational Requirements

- Graceful API, Socket, Consumer, and Cron shutdown.
- Socket readiness verifies Redis subscription; API, Consumer, and Cron readiness verify their required dependencies.
- Readiness fails when a process cannot reach a dependency required for its role.
- Consumer replicas scale independently with Kafka partitions.
- Cron starts with one replica; every job also uses a distributed lock and idempotent database transition.
- Health indicates process liveness only.
- Database migrations are versioned per service and run explicitly against the
  owning service schema. Identity migrations live under
  `be/services/identity/internal/storage/postgres/schema/migrations` and target the `identity` schema.
- Retry uses bounded attempts and exponential backoff.
- Poison events move to a dead-letter path or visible failed state.
- Secrets come from environment or deployment secret storage.
