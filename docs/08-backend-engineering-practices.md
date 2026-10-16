# Backend Engineering Practices

## Purpose

BadmintonHub backend code must be correct, easy to change, observable, and recoverable in production. Correct booking, payment, tenant isolation, and financial data take priority over premature optimization or abstraction.

These practices apply to the API, Socket, Consumer, and Cron processes.

## Core Practices

### 1. Protect correctness first

- Define inputs, outputs, invariants, edge cases, and failure cases before implementation.
- Enforce critical invariants in PostgreSQL as well as application code.
- Treat PostgreSQL as the source of truth for bookings, payments, and financial state.
- Optimize only after measuring a real bottleneck.

For BadmintonHub, correctness includes preventing overlapping bookings, preserving organization isolation, and applying each payment or event once.

### 2. Keep API contracts explicit

- Validate request data at the transport boundary.
- Keep response shapes, stable error codes, pagination, and status mappings consistent.
- Use timeouts for downstream calls.
- Require an idempotency key for retryable write operations where duplicate effects are unsafe.
- Never accept authoritative prices or organization scope from the frontend.

### 3. Separate responsibilities

Follow the existing flow:

```text
router -> handler -> domain service -> repository
```

- Handlers parse and validate transport data.
- Domain services enforce business rules and coordinate workflows.
- Repositories own database access.
- Infrastructure adapters own Redis, Kafka, payment-provider, and other external integrations.
- Business rules must not depend on Gin, SQL rows, or Kafka message types.

Prefer a small, explicit module over a generic abstraction created before a second use case exists.

### 4. Design database behavior deliberately

- Use constraints to protect uniqueness, references, valid states, and non-overlapping bookings.
- Keep transactions as short as possible while preserving atomicity.
- Create indexes from actual query patterns and verify important queries with `EXPLAIN ANALYZE`.
- Avoid N+1 queries and unbounded list queries.
- Use backward-compatible migrations for rolling deployments.
- Never edit generated sqlc code; update migrations or query files and regenerate.

### 5. Handle concurrency explicitly

A read followed by a write is not safe by itself when requests can run concurrently.

Choose an appropriate correctness boundary:

- PostgreSQL constraint for uniqueness or booking overlap.
- Atomic conditional update for state transitions.
- Row lock when several related values must be checked and changed together.
- Optimistic version check when conflicts are uncommon.

Redis locks may reduce contention but must not be the only protection against double booking.

### 6. Make side effects idempotent

Payment callbacks, Kafka messages, Cron jobs, and retryable API commands may run more than once.

- Give each operation a stable identity.
- Record successful processing in the same transaction as the state change where possible.
- Return the previous result when an idempotency key is safely reused.
- Reject reuse when the same key is sent with a different payload.
- Make duplicate delivery observable without treating it as a second success.

### 7. Treat failure as normal

- Every network call must have a timeout.
- Retry only transient failures.
- Use bounded attempts, exponential backoff, and jitter.
- Do not retry validation, authorization, conflict, or other permanent failures.
- Send poison events to a dead-letter path or a visible failed state.
- Use the transactional outbox when a database change and event publication must be atomic.

### 8. Build observability into each flow

Use structured logs, metrics, and traces where they help answer:

- What operation failed?
- Which request, booking, organization, or event was involved?
- Where was time spent?
- Can the operation be retried safely?

Propagate `request_id` and `correlation_id`. Include domain identifiers only when relevant. Never log passwords, tokens, secrets, or complete payment payloads.

### 9. Secure every boundary

- Authentication proves identity; authorization proves permission.
- Apply organization and venue scope in every owner or staff data-access path.
- Use least-privilege credentials for applications and deployment jobs.
- Use parameterized queries.
- Keep secrets in environment or deployment secret storage.
- Avoid exposing internal errors or sensitive personal and payment data.

### 10. Test according to risk

- Unit-test domain rules and state transitions.
- Integration-test PostgreSQL constraints, transactions, Redis behavior, and Kafka idempotency.
- End-to-end test critical business journeys.
- Include concurrent requests, duplicate callbacks, retries, timeouts, and partial failures.
- Prefer real infrastructure containers when correctness depends on infrastructure behavior.

### 11. Deploy backward-compatible changes

Use the expand, migrate, contract sequence:

1. Expand the schema or contract so old and new code can both work.
2. Deploy the compatible application change.
3. Migrate or backfill data safely.
4. Remove the old path only after it is no longer used.

Application rollback changes image versions only. Do not automatically roll back database migrations.

### 12. Keep changes reviewable

- Use clear names and focused functions.
- Keep pull requests small enough to reason about.
- Avoid mixing refactoring with behavior changes unless necessary.
- Document why a non-obvious decision exists.
- Validate configuration at startup and fail with an actionable error.
- Update the relevant contract or design document when behavior changes.

## Backend Review Checklist

Before merging a backend change, verify:

- [ ] Inputs, authorization, organization scope, and error mapping are correct.
- [ ] Running the request, callback, message, or job twice is safe.
- [ ] Concurrent requests cannot violate a business invariant.
- [ ] Database constraints, transactions, and indexes match the access pattern.
- [ ] Downstream calls have timeouts and an appropriate retry policy.
- [ ] Partial failure between the database and external side effects is handled.
- [ ] Logs and metrics are sufficient to diagnose production failure.
- [ ] Secrets and sensitive data are not logged or exposed.
- [ ] Old and new application versions can run during deployment.
- [ ] Unit, integration, and end-to-end coverage matches the risk.
- [ ] Rollback or recovery steps are understood.

## Priority for BadmintonHub

When trade-offs are required, use this order:

1. Data integrity and tenant isolation.
2. Booking and payment concurrency correctness.
3. Idempotency and failure recovery.
4. Clear API and event contracts.
5. Observability and operational readiness.
6. Performance based on measurements.
7. Abstraction and code reuse.
