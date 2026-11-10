---
name: backend
description: Implement, review, and debug the BadmintonHub Go backend across service boundaries, PostgreSQL, Kafka, Redis, and OpenTelemetry.
---

# BadmintonHub Backend

Use this project role for backend implementation and review. Load repository context before changing behavior.

## Context loading

Before writing or editing code, always read `docs/09-project-conventions.md` first.
Treat its naming, package, cSpell, test-location, and folder-ownership rules as
mandatory. Then load the remaining repository context below.

Read `docs/ai-context/INDEX.md`, `docs/08-backend-engineering-practices.md`, `docs/03-system-design.md`, `docs/04-data-model.md`, `docs/05-api-and-events.md`, `docs/design/05-folder-structure.md`, and `docs/design/06-microservice-migration.md`. Then inspect the relevant Go module, contract, migration, configuration, tests, Dockerfile, and Compose path.

## Boundaries

- `be/platform` contains infrastructure lifecycle and cross-cutting primitives; it must not own business rules.
- Each service owns its Go module, configuration namespace, database, migrations, and state. Services must not import one another.
- Gateway is the public HTTP entry point. Identity owns users, credentials, sessions, token issuance, and `/me`.
- Use versioned gRPC/Protobuf for synchronous service calls and versioned Kafka events for asynchronous workflows.
- PostgreSQL owns authoritative business invariants. Redis is an optimization or realtime transport.
- Kafka delivery is at-least-once; consumers must be idempotent and database-to-Kafka flows should use an outbox boundary.

## Implementation rules

- Trace contract -> route -> handler -> service/usecase -> repository/client -> schema/migration -> tests -> deployment.
- Derive actor identity from a verified JWT. Never trust client-supplied user, organization, or venue IDs for authorization.
- Propagate context, deadlines, cancellation, request/correlation IDs, and OpenTelemetry trace context through HTTP and Kafka.
- Define transaction boundaries, duplicate handling, retry policy, partial-failure behavior, and stable error mapping.
- Add focused tests for success, validation, authorization, dependency failure, duplicate/retry behavior, and concurrency risks.
- Never log secrets, tokens, passwords, or complete sensitive payment payloads.
- Report source, test, runtime, deployment, and business verification separately.

## Additional project checks

- Keep booking prices server-calculated and snapshot the value used for a booking or payment.
- Regenerate SQL and contract artifacts after changing their sources; run the relevant drift checks.
