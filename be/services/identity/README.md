# Identity Service

Identity owns user and staff accounts, credentials, sessions, token issuance,
roles, permissions, tenant organizations, venues, memberships, and venue staff
assignments. It owns the `identity` PostgreSQL schema and migration history.
Gateways and Catalog call its versioned gRPC contracts rather than importing
its Go packages or accessing its schema.

## Layout and ownership

```text
identity/
  cmd/identity/main.go              # config, signals, app.NewServer/Run
  config/                           # IDENTITY_* configuration
  internal/
    app/                            # database lifecycle and dependency composition
    constant/                       # resource-prefixed status, role, action, query constants
    domain/
      auth/                         # authentication, sessions, shared token issuer
      user/                         # user CRUD, registration, User portal admission
      staff/                        # staff CRUD, account access, memberships, venue assignments
      tenant/                       # organization CRUD, enrollment, registration review
      venue/                        # venue CRUD and scoped access
      role/                         # role CRUD and permission association
      permission/                   # permission CRUD
    exception/                      # resource errors and gRPC/PostgreSQL mappings
    model/                          # shared entities, session, token pair, list query
    repository/postgres/
      <module>/                     # persistence adapters implementing domain ports
      helper/                       # SQLC projections, nullable values, scoped ID parsing
    security/                       # password hashing and comparison
    validation/                     # shared struct validator
    utils/                          # dependency-free text normalization and syntax checks
    storage/postgres/
      schema/migrations/            # versioned schema changes
      schema/seed/                  # local development data only
      query/                        # handwritten SQL by resource
      generated/sqlc/               # generated code, never edited manually
      transaction/                  # begin, commit, rollback mechanics
    transport/grpc/
      <module>/                     # resource server, handler, and response adapters
      helper/                       # shared response and list-query mapping
      authorization.go              # shared current-actor checks
  test/                             # unit, gRPC, and PostgreSQL integration tests
```

Use `<module>_service.go`, `<module>_input.go`, and `<module>_repository.go`
inside each domain. Use `<module>_server.go`, `<module>_handler.go`, and
`<module>_response.go` in transport. Auth uses actor prefixes: `user_*`,
`tenant_*`, and `admin_*`. Constants use `<resource>_constant.go`; shared
adapter helpers use `<concern>_helper.go`.

Domain services own eligibility, transitions, and workflow order. Repository
interfaces belong to their domain; PostgreSQL adapters implement them. For
multi-record workflows, services execute domain transaction ports while the
adapter handles locks, conditional SQL, mapping, and transaction lifecycle.

## Shared utilities and helpers

- `utils/string_utils.go` normalizes case-insensitive text and checks text/slug
  syntax. The caller supplies resource length limits and chooses its error.
- `repository/postgres/helper` centralizes optional text/timestamp conversion,
  auth-session parameters, account projections, and tenant scope parse errors.
  SQLC types stay inside this adapter boundary.
- `transport/grpc/helper` maps shared user/staff/membership response fields and
  common list filters. Endpoint-specific access fields and timestamp precision
  remain explicit in each resource's response adapter.
- Auth's shared `auth_handler.go` handles User/Tenant session RPCs. Their
  separate servers retain their own injected portal admission policies.
- RPC handlers use a flat flow: authorize, bind input, call the service, map
  errors, and return a response helper. Simple handlers do not need a callback
  wrapper; keep list/response mapping in the resource response file.
- `auth.TokenManager.NewSession` prepares tokens and a hashed session model.
  The calling service decides admission and persistence. Staff revocation
  shares one domain operation for both removal and inactive status updates.

Keep tenant approval rules, staff admission, portal policies, and transaction
orchestration in domain services. Utilities must not import domain, repository,
transport, or Protobuf packages. A helper needs actual consumers; avoid generic
CRUD abstractions that hide resource behavior.

## Authentication and access

User registration belongs to `domain/user`; staff account/password behavior
belongs to `domain/staff`. Auth owns the common credential/session/token flow.
User and Tenant auth share protocol handling but always apply their configured
portal policy. Admin auth remains a separate contract and staff account flow.

Active tenant staff can read their organization's venues. Active venue staff
can read only assigned venues. Owners manage membership and assignments under
Staff, and create venues only for approved, active organizations. Removing or
deactivating a membership revokes its venue assignments in the same transaction.
Adding by email uses an existing registered account; it does not send an invite.

Tenant enrollment records contact details and an idempotent creation result.
The Admin review queue accepts only pending registrations for approval or
rejection; the decision and audit record commit atomically. Migration `000005`
adds contacts and review history. Migrations run as explicit deployment steps.

Identity exposes dedicated Auth, TenantAuth, AdminAuth, User, Staff, Tenant,
Venue, Role, and Permission gRPC services. Callers use the owning resource
contract directly. Coordinate Identity, Gateway, and Catalog releases when
rolling out the resource-service split.

## Errors and verification

`internal/exception` owns sentinels, typed violations, configuration errors, and
resource-specific gRPC/PostgreSQL mappings. It stays independent of domain,
model, repository, and transport packages. See the shared
[project conventions](../../../docs/09-project-conventions.md).

Run `go test ./...` and `go vet ./...` from this module. Run
`bash be/tests/e2e/identity/tenant.sh` from the repository root for race-enabled
Identity and User Gateway tests with disposable PostgreSQL, migrations, and no
development seeds. Integration tests require `TEST_IDENTITY_DATABASE_URL` and
otherwise skip; a unit-only run does not prove database behavior.
