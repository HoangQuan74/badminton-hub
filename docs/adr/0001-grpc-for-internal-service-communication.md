---
status: accepted
---

# Use gRPC for internal service communication

BadmintonHub uses HTTPS/JSON at the User and Admin Gateway boundaries and WSS at
the Socket Gateway boundary. Synchronous business calls between gateways and
backend services, and between backend services, use versioned Protocol Buffer
contracts over gRPC. Asynchronous cross-service workflows use versioned Kafka
events. The frontend never calls a service directly, services do not query
another service's schema, and internal gRPC calls must carry
deadlines, request/correlation IDs, authenticated service identity, and bounded
retry behavior. Docker health probes may keep using HTTP health endpoints for
operational compatibility.

## Consequences

- Protobuf definitions live under `contracts/proto/` and are evolved additively;
  breaking changes require a new package or version.
- The first concrete contract is `contracts/proto/identity/v1/auth.proto`, which
  owns the Gateway-to-Identity Auth calls behind the public User Gateway routes.
- Each service owns its generated gRPC client/server code and exposes only the
  methods needed by another service.
- gRPC deadlines, status codes, metadata propagation, and tracing become part
  of the shared service contract.
- Commands must remain idempotent before a retry is allowed; the client must
  treat timeout results as unknown and reconcile through the owning service.
- The current Phase 1 HTTP reverse proxy is legacy migration code only. It is
  not the service contract and must not be extended with new business routes.
  User, Admin, and future domain gateways must call the owning service through
  its versioned gRPC contract before the HTTP bridge is removed.
- Each stateful service owns a logical PostgreSQL boundary, migration history,
  and namespaced database/Redis configuration. Initial environments may
  co-locate service schemas in one PostgreSQL database, while credentials and
  repository boundaries prevent cross-service schema access. A service can move
  to its own database without changing its gRPC or event contracts.

## Considered Options

- **Private HTTP/JSON:** easy to inspect and replay, but weaker generated
  contracts and less consistent cross-service evolution.
- **gRPC for synchronous calls:** selected for typed contracts, deadlines,
  streaming support, and language-neutral code generation.
- **Kafka for every interaction:** rejected for request/response authorization,
  reads, and commands that need an immediate result; Kafka remains the async
  boundary.

## Current implementation gap

The repository currently contains an HTTP migration bridge so the existing
public routes remain runnable while the gRPC cutover is built. This bridge is
not an exception to the architecture and must not receive new domain behavior.
The cutover blockers and required evidence are tracked in the migration design;
until they pass, source-level HTTP bridge code is deliberately retained for
rollback compatibility.
