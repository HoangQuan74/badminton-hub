# Backend Folder Structure

```text
+ badminton-hub/
  + be/
    + gateways/
      + user/
        + go.mod/go.sum              # User Gateway module boundary
        + cmd/user-gateway/          # Public user Gateway binary
      + admin/
        + go.mod/go.sum              # Admin Gateway module boundary
        + cmd/admin-gateway/         # Private admin Gateway binary
      + socket/
        + go.mod/go.sum              # Socket Gateway module boundary
        + cmd/socket-gateway/        # WebSocket Gateway binary
    + services/
      + identity/
        + go.mod                    # Identity module boundary
        + go.sum
        + cmd/identity/             # Identity service binary
        + internal/
          + domain/
            + auth/                # Credentials, sessions, tokens
            + user/                # User CRUD rules and inputs
            + staff/               # Staff CRUD and access rules
          + model/                 # Reusable Identity entities and query models
          + security/              # Identity-only password helper
          + validation/            # Identity-only email helper
          + storage/postgres/
            + schema/
              + migrations/        # Identity database migrations only
              + seed/              # Identity-only local development seed
            + query/               # Identity SQL source split by module
            + generated/sqlc/      # Identity generated PostgreSQL code
            + repository/
              + auth/              # Auth/session persistence adapter
              + user/              # User persistence adapter
              + staff/             # Staff/access persistence adapter
            + transaction/         # Shared transaction lifecycle
          + transport/grpc/auth/   # Identity Auth gRPC server
          + transport/grpc/user/   # Identity User CRUD gRPC server
          + transport/grpc/staff/  # Identity Staff CRUD gRPC server
        + test/                    # Identity-owned tests and fixtures
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
      + database/                  # PostgreSQL, MongoDB, and Redis lifecycle
      + http/
        + context/                 # Request/correlation metadata
        + server/                  # HTTP lifecycle and graceful shutdown
      + grpc/
        + client/                  # Shared gRPC dialing and client metadata
        + endpoint/                # Shared typed endpoint lifecycle
        + error/                   # Shared gRPC status/error mapping
        + metadata/                # Request and service metadata propagation
        + server/                  # gRPC lifecycle and graceful shutdown
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
  internal domain, transport, storage queries, service-specific tests, and its
  `go.mod/go.sum` stay under that service.
- `gateways/user`, `gateways/admin`, and `gateways/socket` are independent edge
  processes. They own forwarding, edge readiness, and protocol-specific
  boundary behavior; they do not own business tables or domain state.
- `services/identity` owns users, credentials, refresh sessions, token issuance,
  and current-user gRPC methods. Other services reference Identity by its
  versioned gRPC contract and user ID.
- `catalog`, `booking`, `payment`, and `notification` are separate future
  boundaries. Empty directories document ownership; they are not deployed until
  their handlers, invariants, storage, and operational checks exist.
- `consumer`, `cron`, and `socket` are process boundaries with their own runtime
  responsibilities. They must not import another service's transport package.
- `platform` contains infrastructure lifecycle and cross-cutting primitives. It
  must not contain organization, booking, payment, or identity business rules.
- Each service owns a namespaced configuration contract (`IDENTITY_*`,
  `GATEWAY_*`, and later `CATALOG_*`, `BOOKING_*`, `PAYMENT_*`). A service must
  not read another service's schema or Redis configuration.
- Each stateful service owns a logical PostgreSQL boundary, credentials, schema,
  and migration directory. Initial environments may co-locate service schemas
  in one PostgreSQL database, but services must not query another service's
  schema or tables. A service can move to its own database without changing its
  domain or contract boundary.
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
gateways/user/cmd/user-gateway -----> gateways/user -----> platform + contracts
gateways/admin/cmd/admin-gateway ---> gateways/admin ---> platform + contracts
gateways/socket/cmd/socket-gateway -> gateways/socket -> platform + contracts
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
name remains for deployment compatibility; it now builds
`gateways/user/cmd/user-gateway`.
