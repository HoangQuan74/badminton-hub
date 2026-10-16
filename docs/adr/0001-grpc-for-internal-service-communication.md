---
status: accepted
---

# Use gRPC for internal service communication

BadmintonHub uses HTTP/JSON only at the public Gateway boundary. Synchronous
business calls between backend services use versioned Protocol Buffer contracts
over gRPC, while asynchronous cross-service workflows continue to use versioned
Kafka events. The Gateway remains the only public HTTP entry point, services do
not query another service's database, and internal gRPC calls must carry
deadlines, request/correlation IDs, authenticated service identity, and bounded
retry behavior. Docker health probes may keep using HTTP health endpoints for
operational compatibility.

## Consequences

- Protobuf definitions live under `contracts/proto/` and are evolved additively;
  breaking changes require a new package or version.
- Each service owns its generated gRPC client/server code and exposes only the
  methods needed by another service.
- gRPC deadlines, status codes, metadata propagation, and tracing become part
  of the shared service contract.
- Commands must remain idempotent before a retry is allowed; the client must
  treat timeout results as unknown and reconcile through the owning service.
- The current Phase 1 Gateway-to-Identity HTTP/JSON proxy is a migration bridge
  for public-route compatibility. It must be replaced by the Identity gRPC
  contract before adding the next synchronous domain-service dependency.
- Each service also owns namespaced database and Redis configuration; transport
  isolation does not grant a service access to another service's storage.

## Considered Options

- **Private HTTP/JSON:** easy to inspect and replay, but weaker generated
  contracts and less consistent cross-service evolution.
- **gRPC for synchronous calls:** selected for typed contracts, deadlines,
  streaming support, and language-neutral code generation.
- **Kafka for every interaction:** rejected for request/response authorization,
  reads, and commands that need an immediate result; Kafka remains the async
  boundary.
