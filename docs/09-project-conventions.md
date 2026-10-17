# Project Conventions

## Purpose and authority

This is the shared engineering convention for BadmintonHub. It applies to
backend, frontend, DevOps, BA, and QC work.

An accepted ADR or a service README may add detail for its owner, but it must
not contradict this document. When a design changes, update the nearest ADR,
contract, or service README together with the implementation.

The project separates verified facts from assumptions. Completion reports must
identify whether a statement is supported by source code, automated checks,
runtime evidence, deployment evidence, or business verification.

## Architecture and ownership

The production request paths are:

```text
Frontend -- HTTPS/JSON --> User Gateway  -- gRPC/Protobuf --> Service --> logical own storage
Frontend -- HTTPS/JSON --> Admin Gateway -- gRPC/Protobuf --> Service --> logical own storage
Frontend -- WSS --------> Socket Gateway -- gRPC/Protobuf --> Service when required
                                                           \--> Kafka events when asynchronous work is required
```

- User Gateway, Admin Gateway, and Socket Gateway are the frontend entry points.
  They own edge protocol handling, authentication, forwarding, readiness, and
  public API documentation.
- The frontend never calls a service directly. User/Admin use HTTPS/JSON and
  Socket uses WSS/WebSocket.
- Services own business rules, state transitions, authorization decisions,
  schemas, migrations, repositories, and service-level tests.
- Synchronous service calls use versioned Protobuf contracts over gRPC.
- Asynchronous cross-service workflows use versioned Kafka events.
- A service must never import another service module or query another service's
  schema. Use gRPC or Kafka instead.
- `be/platform` owns infrastructure lifecycle and cross-cutting mechanics. It
  must not own identity, booking, payment, catalog, or organization rules.
- PostgreSQL is the authoritative source for business state. Redis is an
  optimization or realtime transport and cannot be the only correctness
  boundary.
- Each stateful service owns a logical database boundary, schema, credentials,
  and migration history. Initial deployments may co-locate those schemas in one
  PostgreSQL database for operational simplicity, but services never access
  another service's schema. Moving a service to its own database is an
  infrastructure change, not a domain or contract change.

The Gateway-to-Service boundary is also a trust boundary. The gateway validates
the public request, while the owning service validates the gRPC input again and
enforces the business invariant before changing state.

## Repository layout

```text
be/
  platform/                         # shared infrastructure-only Go module
  gateways/<gateway>/               # independent public edge module
    cmd/<gateway>/                  # process entrypoint
    config/                         # gateway configuration
    transport/http/                 # public HTTP routes and adapters
    docs/                           # gateway OpenAPI/Swagger artifacts
  services/<service>/               # independent service module
    cmd/<service>/                  # process composition and lifecycle
    config/                         # namespaced service configuration
    internal/
      domain/<module>/              # business rules and module contracts
      model/                        # service-local reusable entities/query models
      security/                     # service-owned security helpers
      validation/                   # service-owned validation primitives
      storage/postgres/
        schema/
          migrations/               # versioned up/down migrations
          seed/                     # local development data only
        query/                      # handwritten SQL split by module
        generated/sqlc/             # generated code; never edit manually
        repository/<module>/        # module-owned persistence adapters
        transaction/                # shared transaction lifecycle
      transport/grpc/<module>/     # private gRPC transport by module
    test/                           # service-owned tests
  tests/
    integration/<owner>/            # cross-package or infrastructure tests
    contract/<area>/                # HTTP, gRPC, and event contract tests
    e2e/<journey>/                  # multi-service user journeys
contracts/
  openapi/                          # public Gateway HTTP contracts
  proto/                            # versioned internal gRPC contracts
  events/                           # versioned Kafka event contracts
docs/                               # requirements, design, ADRs, and practices
devops/                             # Compose, CI/CD, and operational assets
fe/                                 # React/Vite frontend
```

Keep new code inside the owner that owns the behavior. Put code in
`be/platform` only when it is infrastructure-oriented and has more than one
real consumer.

## Dependency direction

```text
gateway router
  -> gateway handler/adapter
  -> service gRPC client
  -> service transport handler
  -> domain service/use case
  -> repository interface
  -> PostgreSQL adapter/sqlc
```

Handlers own transport binding, authorization orchestration, and response
mapping. Domain services own business rules and state transitions. Repository
interfaces belong to the owning domain module; PostgreSQL adapters implement
those interfaces and map SQLC rows to service-local models.

Entry points compose dependencies. They do not contain business rules.

## Naming and spelling

- Use `AnhBan` (PascalCase) for exported Go identifiers: types, functions,
  methods, constants, and exported fields.
- Use `anhBan` (camelCase) for unexported identifiers and local variables.
- Preserve word boundaries and common initialisms: `identityConfig`,
  `gatewayHTTP`, `gatewayProxy`, `HTTPServer`, `URL`, `ID`, `JWT`, and `API`.
- Use lowercase for Go package names and directories. Established compound
  infrastructure package names remain lowercase and concatenated:
  `grpcclient`, `grpcendpoint`, `grpcerror`, `grpcmetadata`,
  `grpcserver`, `httpcontext`, and `httpserver`.
- Import aliases use camelCase with initialisms preserved, such as
  `identityGRPC`, `authPostgres`, and `grpcError`.
- Use `UPPER_SNAKE_CASE` for environment variables.
- Add valid project terms to `.cspell.json`. Do not rename a correct identifier
  only to silence cSpell.

## Model and input organization

Use one service-local domain model when the same business entity crosses domain,
repository, and transport layers. Keep it under `internal/model`.

```text
internal/model/
  user.go                       # reusable User entity
  staff.go                      # reusable Staff entity
  list_query.go                 # reusable pagination/search/date query
```

Module-owned use-case inputs stay with the module:

```text
internal/domain/<module>/
  input.go                      # module use-case inputs and validation tags
  output.go                     # module-owned output types
  repository.go                 # repository interface and persistence params
  service.go                    # business logic
```

Do not create duplicate domain entities merely because two responses have
different shapes. Reuse `model.User` and map it to separate Protobuf response
types or response mappers when transport semantics differ.

Do not share Go business models between services. Cross-service data types are
defined in Protobuf or event contracts and generated per owner.

SQLC row types remain inside `storage/postgres/generated/sqlc`. They are storage
models, not domain models, and must not leak into domain services or handlers.

## Validation, normalization, and business rules

Input structs carry validation tags. Each service exposes one shared validator
through `internal/validation.Struct`.

```go
if err := validation.Struct(input); err != nil {
    return ErrInvalidInput
}

// Canonicalize values after validation.
input.Email = strings.ToLower(strings.TrimSpace(input.Email))
```

Follow this order:

1. Validate the complete input struct once.
2. Canonicalize or normalize values without performing a second validation.
3. Apply business rules in the domain service.
4. Enforce authoritative invariants with database constraints or atomic writes.

`validator.Struct` checks data but does not mutate it. Canonicalization such as
trimming whitespace or lowercasing email remains an explicit step. Transport
binding must not be the only validation boundary.

## PostgreSQL storage

Every stateful service owns its schema and migration history. Identity uses the
following layout:

```text
internal/storage/postgres/
  schema/
    migrations/                   # schema changes with up/down files
    seed/                         # local development seed only
  query/
    auth.sql                      # auth and session SQL
    user.sql                      # user CRUD SQL
    staff.sql                     # staff/access SQL
  generated/sqlc/                 # generated pgx code
  repository/
    auth/repository.go
    user/repository.go
    staff/repository.go
  transaction/transaction.go
```

- Write SQL only in `query/<module>.sql`.
- Run the service SQLC generator after changing SQL or schema.
- Never edit `generated/sqlc` manually.
- Repositories expose domain-relevant operations and return `internal/model`
  values, never SQLC rows.
- Keep transaction ownership at the repository/use-case boundary that needs
  atomicity. Use the shared transaction helper for lifecycle and rollback.
- Put duplicate handling, constraints, locking, and state invariants in the
  database where the database can enforce them safely.
- Migrations must be backward-compatible with the previous application version
  during rolling deployment when the service is deployed incrementally.

## gRPC transport

Each private gRPC module uses:

```text
internal/transport/grpc/<module>/
  server.go
  errors.go
  response.go
```

- `server.go` contains generated server methods, constructors, authorization
  calls, domain calls, and typed response mapping.
- `errors.go` maps domain errors through shared `be/platform/grpc/error` rules.
- `response.go` maps service-local models to generated Protobuf responses,
  including list responses.
- Use `be/platform/grpc/endpoint.Run` for the common authorization, execution,
  error mapping, and typed response flow.
- Keep authorization before domain execution. Derive actor identity from a
  verified token and enforce permissions in the owning service.
- Keep generated response types module-specific. Do not use
  `map[string]any` or `google.protobuf.Any` as a shortcut.
- Evolve Protobuf contracts additively. A breaking change requires a new
  versioned package.

Platform owns gRPC lifecycle, metadata, status mapping, and endpoint mechanics.
Business permissions and response semantics remain in the service.

## Code formatting and reviewability

Format code for human review, not only compiler acceptance.

- Keep one meaningful operation per visual block.
- Start each meaningful processing block with one short English comment that
  names its purpose: `Validate input`, `Canonicalize values`, `Persist the
  session`, or `Map the response`.
- Do not comment every line or restate syntax. Comments explain intent and step
  boundaries.
- Keep validation, canonicalization, authorization, business execution,
  persistence, and response mapping in separate blocks.
- Use one local variable and one error check for each meaningful operation.
- Expand callbacks, nested structs, composite literals, and calls with more
  than two meaningful arguments.
- Avoid dense one-line handlers, chained conditionals, and anonymous blocks that
  hide control flow.
- Use explicit names such as `accessToken`, `userParams`, `staffRepository`,
  and `userResponse` when they clarify context.
- Run `gofmt`; remember that `gofmt` alone does not guarantee readable code.

## API, event, and reliability rules

- Trace contract → route → handler → service/use case → repository/client →
  schema/event → tests → deployment.
- Use stable error codes and documented status mappings.
- Validate at the transport boundary and enforce authorization, tenant scope,
  and business invariants in the owning service.
- Commands with unsafe duplicate effects require an idempotency strategy before
  retries are enabled.
- Treat timeouts as unknown outcomes. Do not assume a remote write did not
  happen; reconcile through the owning service when necessary.
- Propagate deadlines, cancellation, request/correlation IDs, authenticated
  service identity, and OpenTelemetry context through HTTP, gRPC, and Kafka.
- Never log passwords, tokens, secrets, or complete sensitive payloads.
- Keep local state changes and external side effects separate. Use an outbox,
  inbox, reconciliation, or compensation when one transaction cannot cover
  both.

## Test conventions

Tests belong to the owner that owns the behavior:

```text
be/platform/test/
be/services/<service>/test/
be/tests/integration/<owner>/
be/tests/contract/<area>/
be/tests/e2e/<journey>/
fe/tests/
```

- Do not place new test files beside production source.
- Domain rules and state transitions require unit tests.
- SQL constraints, transactions, and repository behavior require PostgreSQL
  integration tests.
- API and event shape changes require contract tests.
- Auth, duplicate callbacks, expiry, authorization, retries, timeouts, and
  concurrency require explicit failure-path coverage.
- Tests skipped because infrastructure is unavailable are `unverified`, not
  passed.
- Record the command, prerequisites, expected result, actual result, and
  evidence layer for important checks.

## Documentation, change, and completion

- Update the nearest contract, ADR, README, or design document when behavior or
  ownership changes.
- Keep a change focused. Do not mix unrelated refactors into a feature.
- Review correctness, ownership, authorization, concurrency, idempotency,
  failure handling, observability, tests, migration, and rollback before style.
- For migrations and contract changes, document compatibility, rollout order,
  rollback or compensation, and cleanup conditions.
- Completion reports separate:
  - source changes
  - automated checks
  - compatibility/parity
  - deployment and rollout
  - runtime and business verification
  - remaining risks and next actions
