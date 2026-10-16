# Backend Services

Each directory under `services/` is one ownership and deployment boundary.
Service code owns its transport, domain rules, adapters, storage queries, and
service-specific tests. Shared infrastructure belongs in `platform/`.

Each deployable service is also an independent Go module with its own
`go.mod`/`go.sum`. A service may import the infrastructure-only `platform`
module, but it must never import another service module. Identity-only helpers
live under `services/identity/pkg` and are not shared with Gateway.

Each service owns a namespaced configuration contract. For example, Identity
uses `IDENTITY_DATABASE_URL`, `IDENTITY_REDIS_URL`, `IDENTITY_JWT_SECRET`, and
Identity-specific timeout/TTL variables; Gateway uses `GATEWAY_*`; process
boundaries use their own prefixes such as `CONSUMER_REDIS_URL` and
`SOCKET_REDIS_URL`. A service does not read another service's database or Redis
variables.

Each service also owns its database and migration history. Identity migrations
live in `services/identity/migrations` and run only against
`IDENTITY_DATABASE_URL`; a future service adds its own `<service>/migrations`
directory and database URL. A PostgreSQL server may host several service
databases, but business tables are never shared across those databases.

| Service | Current role | State owner |
| --- | --- | --- |
| `gateway` | Public HTTP compatibility and forwarding | No business state |
| `identity` | Users, credentials, refresh sessions, JWT, `/me` | `users`, `auth_sessions` |
| `catalog` | Organizations, venues, courts, hours, pricing | Future Catalog database |
| `booking` | Holds, bookings, blocks, overlap rules | Future Booking database |
| `payment` | Intents, callbacks, refunds, payouts | Future Payment database |
| `notification` | Notification policy and delivery commands | Future notification state |
| `consumer` | Kafka/outbox consumption and realtime fan-out | No business state |
| `cron` | Scheduled transitions and payout jobs | No duplicate business state |
| `socket` | Authenticated WebSocket connections and local fan-out | No business state |

Synchronous service calls use gRPC/Protobuf contracts under `contracts/proto`.
Asynchronous workflows use versioned Kafka events under `contracts/events`.
