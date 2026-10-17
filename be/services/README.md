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

| Service | Current role | State owner |
| --- | --- | --- |
| `gateway` | Public HTTP compatibility and forwarding | No business state |
| `identity` | Users, credentials, refresh sessions, JWT, `/me` | Identity PostgreSQL database |
| `catalog` | Organizations, venues, courts, hours, pricing | Catalog PostgreSQL database |
| `booking` | Holds, bookings, blocks, overlap rules | Booking PostgreSQL database |
| `payment` | Intents, callbacks, refunds, payouts | Payment PostgreSQL database |
| `notification` | Notification policy and delivery commands | Future notification state |
| `consumer` | Kafka/outbox consumption and realtime fan-out | No business state |
| `cron` | Scheduled transitions and payout jobs | No duplicate business state |
| `socket` | Authenticated WebSocket connections and local fan-out | No business state |

Synchronous service calls use gRPC/Protobuf contracts under `contracts/proto`.
Asynchronous workflows use versioned Kafka events under `contracts/events`.
