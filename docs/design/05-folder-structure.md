# Backend Folder Structure

```text
+ badminton-hub/
  + be/
    + services/
      + gateway/
        + go.mod                    # Gateway module boundary
        + go.sum
        + cmd/gateway/              # Public HTTP Gateway/BFF binary
        + transport/http/          # Public route forwarding and readiness
      + identity/
        + go.mod                    # Identity module boundary
        + go.sum
        + cmd/identity/             # Identity service binary
        + domain/
          + auth/                  # Credentials, sessions, tokens
          + user/                  # Identity user model
        + transport/http/
          + apperror/              # Identity domain-to-HTTP mapping
          + handler/               # Identity HTTP handlers
          + middleware/            # Auth and endpoint middleware
          + i18n/                  # Identity-localized messages
          + models/                # Identity request/response contracts
        + adapters/postgres/       # Identity persistence adapter
        + storage/
          + query/                 # Identity SQL source
          + sqlc/                  # Identity generated PostgreSQL code
        + pkg/
          + helper/                # Identity-only password helper
          + utils/                 # Identity-only email helper
        + migrations/              # Identity database migrations only
        + seed/                    # Identity-only local development seed
      + catalog/
        + go.mod/go.sum              # Catalog module boundary
        + migrations/              # Catalog database migrations
        + domain/
          + organization/          # Future organization ownership
          + venue/                 # Future venue/court configuration
          + availability/          # Future hours/pricing availability
      + booking/
        + go.mod/go.sum              # Booking module boundary
        + migrations/              # Booking database migrations
        + domain/booking/          # Future holds, bookings, blocks, overlap rules
      + payment/
        + go.mod/go.sum              # Payment module boundary
        + migrations/              # Payment database migrations
        + domain/payment/          # Future intents, callbacks, refunds, payouts
      + notification/
        + go.mod/go.sum              # Notification module boundary
        + domain/notification/     # Future notification policy
      + consumer/
        + go.mod/go.sum              # Consumer module boundary
        + notification/            # Kafka notification consumers
        + realtime_dispatcher/     # Kafka-to-Redis fan-out
      + cron/
        + go.mod/go.sum              # Cron module boundary
        + hold_expiration/         # Scheduled hold transitions
        + payout/                  # Scheduled payout work
      + socket/
        + go.mod/go.sum              # Socket module boundary
        + auth/                    # WebSocket authentication
        + client/                  # Connection lifecycle
        + hub/                     # Local connection hub
        + platform/                # Socket-specific adapters
        + subscriber/redis/         # Redis Pub/Sub subscriber
    + platform/
      + go.mod                      # Shared infrastructure module only
      + go.sum
      + config/                    # Environment/config loading
      + database/                  # PostgreSQL connection lifecycle
      + httpcontext/               # Request/correlation metadata
      + httpserver/                # HTTP lifecycle and graceful shutdown
      + redis/                     # Redis client lifecycle
      + mongodb/                   # MongoDB client lifecycle
      + kafka/                     # Kafka client lifecycle
      + logger/                    # Shared logging primitives
      + metrics/                   # Shared metrics primitives
    + Makefile                      # Multi-module local orchestration
  + contracts/
    + openapi/                     # Public Gateway contracts
    + events/                      # Kafka event envelopes and schemas
    + proto/                       # Internal gRPC contracts
  + devops/
  + docs/
```

## Boundary Rules

- `services/<name>` is the ownership boundary for one deployable service. Its
  transport, domain, adapters, storage queries, service-specific tests, and its
  `go.mod/go.sum` stay under that service.
- `services/gateway` owns public HTTP compatibility and forwarding. It does not
  own business tables or domain state.
- `services/identity` owns users, credentials, refresh sessions, token issuance,
  and `/api/v1/me`. Other services reference identity by contract and user ID.
- `catalog`, `booking`, `payment`, and `notification` are separate future
  boundaries. Empty directories document ownership; they are not deployed until
  their handlers, invariants, storage, and operational checks exist.
- `consumer`, `cron`, and `socket` are process boundaries with their own runtime
  responsibilities. They must not import another service's transport package.
- `platform` contains infrastructure lifecycle and cross-cutting primitives. It
  must not contain organization, booking, payment, or identity business rules.
- Each service owns a namespaced configuration contract (`IDENTITY_*`,
  `GATEWAY_*`, and later `CATALOG_*`, `BOOKING_*`, `PAYMENT_*`). A service must
  not read another service's database or Redis configuration.
- Each service owns a separate database and its migration directory. A shared
  PostgreSQL server is allowed, but services must use separate database names,
  credentials, and migration histories; there is no shared business database.
- A service module must never import another service module. `platform` is a
  separate infrastructure-only module; it may be imported by services, but it
  must not contain business rules or service-owned state.
- Public HTTP contracts live under `contracts/openapi`; internal synchronous
  calls use versioned Protobuf under `contracts/proto`; asynchronous workflows
  use versioned Kafka events under `contracts/events`.
- A service must not query another service's tables directly. Use gRPC for
  synchronous calls and Kafka events for asynchronous workflows.

## Dependency Direction

```text
services/gateway/cmd/gateway  ------> services/gateway ------> platform + contracts
services/identity/cmd/identity ------> services/identity -----> platform + identity-local pkg + contracts
services/consumer/cmd/consumer ------> services/consumer -----> platform + contracts
services/cron/cmd/cron -------> services/cron -----------> platform + contracts
services/socket/cmd/socket ------> services/socket ---------> platform + contracts

Frontend ----------> Gateway HTTP/OpenAPI
Service ------------> gRPC/Protobuf or Kafka events
Service ------------> its own storage adapter/schema
```

Entry points only compose dependencies. Service packages own business rules;
platform packages own infrastructure setup. The current `api` container/image
name remains for deployment compatibility, while its source entry point is now
`services/gateway/cmd/gateway`.
