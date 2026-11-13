# Identity tenant CRUD verification

Run from the repository root with Docker, Go, and Python 3 available:

```bash
bash be/tests/e2e/identity/tenant.sh
```

The script starts a uniquely named Compose project with PostgreSQL on a random
loopback port, applies Identity migrations without seeds, and runs Identity and
User Gateway tests with the race detector. It builds/starts Identity, then tests
the real Gateway HTTP adapters through gRPC and PostgreSQL. It stops its Identity
process and removes only its own Compose project on exit. Test logs/binaries
remain in the printed temporary evidence directory.

Expected: exit 0; fresh registration, owner enrollment, replay/conflicts,
tenant list/get/update/archive and isolation, legacy organization-route
compatibility, concurrent key/slug handling, failed-write rollback, account restrictions,
expiration, localized errors, required response fields, and pending-organization
venue denial pass. No existing local stack or deployed database is modified.

To run individual suites, point `TEST_IDENTITY_DATABASE_URL` at a disposable
database migrated through `000004`; otherwise DB tests explicitly skip. Set
`TEST_IDENTITY_GRPC_TARGET` to Identity using that same disposable database for
the Gateway integration journey. Do not use Dev/Production or a shared database.

This establishes local backend behavior. It does not verify frontend flows,
full-stack Compose deployment, public tenant pages, admin review, or booking.

For generation checks, use sqlc **v1.30.0** and the Protobuf tool versions listed
in generated files. Regenerate `tenant.proto` from `contracts/proto` with
`paths=source_relative` into `contracts/gen`, and run `sqlc generate` from the
Identity module. Generation must leave their generated output unchanged.

## Local verification on 2026-11-11

- `bash be/tests/e2e/identity/tenant.sh`: passed against a disposable PostgreSQL
  database migrated without seeds and a real Identity process. Both module
  suites ran with `-race`; HTTP creation/replay/conflicts, tenant CRUD, isolation,
  legacy owner-only creation, rollback and concurrency checks passed. The runner
  removed its own Compose project afterwards.
- `make -C be test`: service/gateway suites passed. Platform lifecycle initially
  failed because the sandbox denied a localhost listener; rerunning
  `go test ./...` in `be/platform` outside the sandbox passed. Empty placeholder
  modules report no packages.
- `make -C be vet` and `make -C be build`: passed.
- Protobuf/sqlc regeneration produced identical output; OpenAPI parsing,
  reference/operation-ID checks and served-spec parity passed.
- Postman CLI v1.74.0 printed no syntax issues, then exited 139 on macOS
  Keychain access (`SecItemCopyMatching failed -50`). Its `status` command is
  unavailable in this version; no Postman workspace verification is claimed.

These checks establish local backend behavior. FE, deployment and public tenant
pages remain outside this verification.

## Exception package refactor verification

On 2026-11-11, all seven affected service/gateway modules passed `go test ./...`
and `go build ./...`; `make -C be vet` passed. Catalog additionally passed
`go test -race ./...`. The isolated `tenant.sh` journey passed again with race
detection against real Identity/PostgreSQL after the error definitions and
mappings moved to their owners' exception packages. A source audit found no
remaining sentinels, error structs, status-error declarations or error mappers
outside those packages, and no reverse imports into business/storage layers.
Public status codes and messages remain unchanged by the refactor.
