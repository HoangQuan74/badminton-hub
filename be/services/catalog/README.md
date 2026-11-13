# Catalog service

Identity owns organizations, venues, staff memberships and venue assignments.
Catalog owns courts, image metadata and the `catalog.v1` gRPC contract. Catalog
calls Identity to authorize each scoped operation; it never reads Identity tables
or trusts the `actor_user_id` request field. Set `CATALOG_IDENTITY_GRPC_TARGET`
to Identity (default `127.0.0.1:9091`; Compose uses `identity:9091`). If Identity
is unavailable, scoped operations fail closed. Authorization and a subsequent
Catalog database write are separate operations, so concurrent revocation is not
atomic with that write.

The local Compose stack runs Catalog on gRPC `:9094`, applies the Catalog
migration and development seed, and exposes public and tenant routes through the
User Gateway.

## Handler and adapter organization

Court, venue and image RPCs live in `court_handler.go`, `venue_handler.go` and
`image_handler.go`. Handlers authorize first, call the domain service directly,
map errors through `internal/exception`, and use `catalog_response.go` for
response projection. Simple calls do not need a nested endpoint callback.

The Identity directory adapter follows the same flat flow. It retains its
five-second RPC deadline (bounded by the caller), deferred cancellation, bearer
token checks and operation-specific dependency/conflict mapping. Response
conversion lives in `directory_response.go`; shared timeout/role constants live
in `internal/constant/identity_constant.go`.

## Errors

`internal/exception` owns Catalog errors, configuration/storage failures, gRPC
status mappings, Identity dependency-error translation, and PostgreSQL conflict
classification. Domain, adapters, repositories and transport use this package;
error definitions are not duplicated in feature folders.

## Database queries and checks

Edit SQL in `internal/storage/postgres/query/` and regenerate the sqlc output
with `make -C be generate` from the repository root. Run `make -C be sqlc-check`
and `cd be/services/catalog && go test ./...` before delivery. Integration tests
require a dedicated disposable PostgreSQL database via `TEST_CATALOG_DATABASE_URL`.

## Catalog endpoints

The User Gateway exposes tenant-scoped organization, venue, and court CRUD, as
well as public venue list/detail. A verified `tenant_owner` can create an
organization; Identity creates its owner membership atomically and sets the
organization to `PENDING_VERIFICATION`. Only pending organizations can be
archived; active organizations require the separate administrative lifecycle.
Organization approval and customer-to-owner onboarding are not implemented by
these endpoints. Tenant owners can create a `DRAFT` venue in an active
organization, edit its fields, and archive it. Tenant staff can read all venues
in their tenant; venue staff can read only assigned venues. Archive
preserves its courts and images and removes the venue from public reads. Approval
and publication are separate workflows; these endpoints do not publish venues.
The complete HTTP contract is `contracts/openapi/user-gateway.yaml`.

Migrations `000002` and `000004` add `ARCHIVED` states. Their down migrations
refuse to run until archived records are explicitly restored to a supported
state. They do not silently rewrite existing data.

## Local seed

The Identity seed creates one active organization, owner/staff memberships and
three venues; the Catalog seed creates six courts. They are for local use and
use `ON CONFLICT DO NOTHING`, so rerunning does not reset edits. See
`devops/README.md` for seed commands and the existing-data cutover procedure.
The owner and staff credentials are documented in `devops/README.md`.

## Venue images

Images use an S3-compatible private bucket. Set `CATALOG_S3_ENABLED=true`,
`CATALOG_S3_BUCKET`, and `CATALOG_S3_REGION`; set `CATALOG_S3_ENDPOINT` for
MinIO or another compatible endpoint. If `CATALOG_S3_ACCESS_KEY` and
`CATALOG_S3_SECRET_KEY` are unset, the AWS SDK default credential chain is used.
For production, provision the bucket and credentials separately. Local Compose
builds MinIO from a pinned upstream source revision and creates the bucket with
its one-shot `minio-init` job. Image storage is disabled by default for a host
process; when disabled, image endpoints return a dependency error.

- `POST /api/v1/tenant/venues/{venue_id}/images`: owner bearer token and
  `multipart/form-data` field `file`. Accepts JPEG/PNG up to 2 MiB and 16
  million pixels. Uploads are not idempotent; retrying can create another image.
- `GET /api/v1/tenant/venues/{venue_id}/images` lists images for active members;
  returned tenant URLs require a bearer token, including for draft previews.
- `GET /api/v1/venues/{venue_id}/images` and
  `GET /api/v1/venue-images/{image_id}` expose images only while the venue is
  published. Responses are not cached, so later archive is respected.

Example, with a local tenant token and seed venue:

```bash
curl -X POST "$BASE_URL/api/v1/tenant/venues/20000000-0000-4000-8000-000000000001/images" \
  -H "Authorization: Bearer $TENANT_TOKEN" \
  -F 'file=@/path/to/venue.jpg'
```

Catalog uploads the object before inserting metadata and attempts to delete the
object if the write fails. A process crash or failed cleanup can leave an orphan;
a reconciliation job is still needed before claiming production-ready recovery.
Archive does not delete images. The `000003` down migration drops image metadata
and does not delete bucket objects; backups and explicit cleanup are required.

## Disposable HTTP E2E

`be/tests/e2e/catalog/smoke.py` exercises the seeded owner/staff accounts,
organization/venue/court operations, and image uploads through the HTTP Gateway.
It writes test data and must only target a disposable environment:

```bash
CATALOG_E2E_ALLOW_WRITES=1 \
CATALOG_E2E_BASE_URL=http://127.0.0.1:18080 \
CATALOG_E2E_PASSWORD='<local-seed-password>' \
python3 be/tests/e2e/catalog/smoke.py
```

The test requires the local seed, migrations, Identity, Catalog, Gateway,
PostgreSQL, and the MinIO bucket to be ready. It does not clean up published
venue images; destroy the disposable test environment after the run.
