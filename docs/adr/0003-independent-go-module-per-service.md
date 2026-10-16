---
status: accepted
---

# Use an independent Go module per service

Each deployable backend service is an independent Go module under
`be/services/<service>`. It owns its `go.mod`, `go.sum`, command entry point,
source packages, tests, and dependency versions. A service must never import
another service's module or package.

Shared infrastructure is isolated in the `be/platform` Go module. Services may
depend on that module for transport, configuration, and connection lifecycle
primitives, but `platform` cannot contain business rules or service-owned
state. Identity-only helpers stay under `services/identity/pkg`.

The root `be/Makefile` may orchestrate module-local test, vet, build, and
migration commands, but `be` is not itself a Go module. Each Docker image
copies and builds only its service module plus the infrastructure module.

## Consequences

- A dependency upgrade or vulnerable package can be changed and released per
  service without changing another service's module graph.
- CI must test each module separately, while the root Makefile remains a local
  convenience wrapper.
- Shared code needs an explicit infrastructure or contract ownership decision;
  it cannot be added to a global business `pkg` directory.
- Cross-service reuse happens through gRPC/Protobuf or Kafka contracts, never Go
  imports.
