# Tenant management delivery

## Confirmed scope

- Group tenant, venue, staff, role and permission by module. Authentication remains under auth.
- Use module-prefixed gRPC server, handler and response filenames.
- Staff owns organization memberships and venue assignments.
- Identity main only loads config, handles signals and calls app.NewServer/Run; app owns storage and composition.
- Remove old aggregate gRPC registrations and duplicate tenant CRUD routes; migrate callers directly.
- Deliver FE CRUD for tenant, venue and staff, plus role/permission management.
- Deliver a simple tenant registration screen with contact details and an admin contact/review queue.
- Approval activates the organization before venue creation; review decisions are atomic and audited.

## Progress

- [x] Tenant file consolidation, separate Venue module and Staff-owned assignments.
- [x] Module server naming and internal/app composition.
- [x] Dedicated permission CRUD and direct resource service registration.
- [x] Initial tenant FE and Admin staff/role/permission CRUD integration.
- [x] Remove duplicate Catalog organization operations and synchronize tenant/Admin HTTP contracts.
- [x] Persist registration contacts and add Admin registration list/approve/reject APIs.
- [x] Deliver tenant venue and membership CRUD APIs/UI.
- [x] Add registration/review screens and API error/empty/loading states in both languages.
- [x] Run focused tests, race-enabled PostgreSQL E2E, FE lint/typecheck/build and direct HTTP CRUD checks.
- [x] Scan Identity authored Go files and consolidate repeated code into shared utilities and layer-owned helpers.
- [x] Verify browser registration, contact display, tenant edit, venue creation, staff reactivation and venue assignment.
- [ ] Verify the Admin approval confirmation dialog in a normal browser session; automated browser control times out on that tab.

## Identity cleanup

The scan covered all authored Go files in `cmd`, `config`, `internal`, and
`test` (117 production files and 17 test files after extraction). Generated
SQLC files remain generator-owned. Exact function-body comparison and source
review identified duplicate account mapping, tenant scope parsing, nullable
values, slug/text validation, auth handlers, token/session preparation, list
filters, admin authorization, and staff revocation.

Shared code now has explicit ownership:

- `internal/utils/string_utils.go`: case normalization and text/slug checks.
- `internal/repository/postgres/helper`: account/session projections, nullable
  PostgreSQL values, and tenant scope parse mapping.
- `internal/transport/grpc/helper`: common response fields and list-query binding.
- Auth domain/transport: common token/session preparation and User/Tenant session
  handlers, with separate injected portal policies.
- Staff domain: one atomic organization/venue revocation flow for removal and
  inactive membership updates.

Tenant creation/review, user registration, permission validation, password
processing, and staff revocation orchestration live in the owning domain
services. The domain migration is incomplete; the source audit below records
remaining policy and workflow ownership work. Thin actor-specific forwarding
methods remain explicit. Constants, including pagination bounds, live in
`internal/constant`. Repository ports consistently name context, identifiers,
and inputs.

### Remaining domain ownership work

- [ ] Move active/admin/tenant-owner admission checks from gRPC authorization
  into the owning domain services; keep token extraction and gRPC error mapping
  at the transport boundary.
- [ ] Express staff organization/venue assignment eligibility in the Staff
  domain. Active owner, matching role, and active membership checks currently
  depend on SQL predicates for the Set operations.
- [ ] Express venue access decisions in the Venue domain. `AuthorizeVenue`
  currently delegates owner/staff/assignment policy to its repository query.
- [ ] Move staff session rotation orchestration into a domain service through
  transaction ports. PostgreSQL currently coordinates lock, revoke, and insert.

Retain tenant-scoped query predicates, conditional writes, row locks, and
database constraints when moving policy decisions. These enforce isolation and
atomicity even after the domain owns the business decision.

## Flat gRPC handler flow

Identity RPC handlers now authorize, bind domain input, call the service, map
errors, and return a response mapper directly. The fixed `grpcEndpoint.Run`
callback wrappers were removed from 53 handlers. Simple calls/signatures stay
compact; structured inputs and list responses remain in their own blocks or
response helpers. Admin Login now has one service call, one error guard, and
one response helper. Resource error mapping, authorization order, enrollment
error details, and safe failure logging are retained.

## Service-wide readability

The flat handler convention applies across `be/services`:

- Identity: 53 unary handler wrappers removed; response projection stays in the
  resource response files.
- Catalog: 18 unary handler wrappers and 8 Identity client callback wrappers
  removed. Explicit timeout/cancellation and error categories are preserved.
  Handler filenames use resource prefixes; directory response conversion is
  separated from RPC execution.
- Socket: existing authentication/subscription/forwarding flow is already flat;
  streaming and JWT library callbacks remain where required.
- Consumer: existing Go code is configuration validation and already uses direct
  control flow.
- Booking, Payment, Notification and Cron: module scaffolds with no Go handlers.
  The shared convention applies to future implementation.

## Verification evidence

- `go test ./...` and `go vet ./...` passed in Identity.
- `bash be/tests/e2e/identity/tenant.sh` passed with race detection against a
  disposable PostgreSQL database and migrations 1 through 5. It exercised real
  Identity and User Gateway contracts, enrollment retries/concurrency/rollback,
  permissions, registration review, venue access, and staff revocation.
  Final flat-handler run evidence: `badminton-tenant-e2e.PMRZG0` under the system temp directory.
- Service-wide `go test -race ./...`, `go vet ./...`, and `go build ./...`
  passed for Identity, Catalog, Socket, and Consumer. Catalog adapter regression
  tests cover bounded/caller deadlines, deferred cancellation, bearer/owner
  scope, dependency failures, and venue conflict categories. Catalog database/S3
  integration tests still require their dedicated dependencies and were skipped
  in this unit/race run.
- Admin Gateway and Catalog module checks passed. Admin Gateway currently has
  no automated tests; its new routes were verified through real HTTP calls.
- Final FE lint, typecheck (included in build), and production build passed,
  including the display/navigation correction and empty Catalog response fix.
- Pre-commit checks passed for backend unit tests and vet, plus FE lint,
  standalone typecheck, and build. Platform lifecycle tests require permission
  to bind localhost and passed outside the sandbox. The platform coverage gate
  remains below 100%: current coverage is 72.2%, compared with 70.1% for an
  isolated copy of HEAD. Both changed platform production files (gRPC error
  mapping and CORS) have 100% function coverage; the existing module-wide
  coverage deficit remains outstanding.
- The isolated HTTP smoke journey passed contact capture, admin approval and
  retry/conflict behavior, venue CRUD, staff assignment/revocation, admin staff
  CRUD, role/permission CRUD, and pending tenant update/archive. Catalog returns
  404 for venue access after revocation, per its public contract.
- Browser checks verified User login, tenant registration submission and success,
  Admin login, and contact display in the pending queue. A fresh Tenant tab
  verified tenant edits, venue creation, staff reactivation and venue assignment.
  Empty court responses now render the empty state instead of a query error.
  Screenshot evidence: `/private/tmp/badminton-management-smoke.mHegDR/tenant-management.jpg`.
  The Admin confirmation tab still has browser-control timeouts; approval was
  independently verified through real HTTP and PostgreSQL integration tests.
- Both OpenAPI source files parse, local references resolve, operation IDs are
  unique, path parameters are complete, and served Gateway copies match.
- Identity SQLC and Identity/Catalog Protobuf artifacts match fresh generator
  output. Unchanged Socket generated files use an older generator version and
  were excluded from this feature's drift verdict.

The installed Postman CLI does not support `postman status`; it was not used as
API verification evidence. Runtime verification used the real local Gateways,
Identity/Catalog services, and disposable PostgreSQL instead.

No deployment or shared development database migration was performed. Disposable
fixtures and PostgreSQL were used for verification; the task-owned test stack
has been stopped.
