# Backend Services

Each directory under `services/` is one ownership and deployment boundary.
Service code owns its transport, domain rules, adapters, storage queries, and
service-specific tests. Shared infrastructure belongs in `platform/`.

Each deployable service is also an independent Go module with its own
`go.mod`/`go.sum`. A service may import the infrastructure-only `platform`
module, but it must never import another service module. Identity-only helpers
live under `services/identity/internal` and are not shared with Gateway.

Each service owns a namespaced configuration contract. For example, Identity
uses `IDENTITY_DATABASE_URL`, `IDENTITY_REDIS_URL`, `IDENTITY_JWT_SECRET`, and
Identity-specific timeout/TTL variables; Gateway uses `GATEWAY_*`; process
boundaries use their own prefixes such as `CONSUMER_REDIS_URL` and
`SOCKET_REDIS_URL`. A service does not read another service's schema or Redis
variables.

Each stateful service owns a logical PostgreSQL boundary, schema, credentials,
and migration history. Identity migrations live in
`services/identity/internal/storage/postgres/schema/migrations` and target the
database selected by `IDENTITY_DATABASE_URL`. Initial environments may
co-locate service schemas in one PostgreSQL database, but business tables and
schema access are never shared across services. A future service adds its own
database connection when it needs physical separation.

| Service | Current implementation | State owner |
| --- | --- | --- |
| `identity` | Accounts, auth, sessions, tenant/venue directory, staff, roles, permissions | Identity PostgreSQL schema |
| `catalog` | Courts, image metadata, scoped directory calls to Identity | Catalog PostgreSQL schema and private object storage |
| `socket` | Authenticated gRPC stream and Redis Pub/Sub fan-out behind Socket Gateway | No durable business state |
| `consumer` | Kafka configuration validation; event handling remains planned | No business state |
| `booking` | Module scaffold for booking and availability rules | Future Booking state |
| `payment` | Module scaffold for intents, callbacks, refunds and payouts | Future Payment state |
| `notification` | Module scaffold for notification policy and delivery | Future Notification state |
| `cron` | Module scaffold for scheduled jobs and locking | No duplicate business state |

## Handler readability

All services follow the [shared project conventions](../../docs/09-project-conventions.md).
Keep unary handlers flat: authorize, prepare domain input, execute the service,
map its error, and return a response mapper. Keep simple calls/signatures compact
and put structured inputs in explicit blocks. Resource handlers use
`<resource>_handler.go`; mapping belongs in response files.

Client adapters also call RPCs directly with explicit deadlines, cancellation,
and dependency-error mapping. Callback APIs remain appropriate for transactions,
streaming/concurrency, lifecycle registration, and SDK configuration. Keep their
bodies small and extract meaningful operations when they grow.

Identity and Catalog use this handler style. Socket's stream already has a
straight authentication/subscription/forwarding flow. Scaffold modules inherit
the convention when implementation is added; no placeholder handlers are needed.

Synchronous service calls use gRPC/Protobuf contracts under `contracts/proto`.
Asynchronous workflows use versioned Kafka events under `contracts/events`.
