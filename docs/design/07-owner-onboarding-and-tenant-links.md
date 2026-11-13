# Tenant CRUD, Owner Enrollment, and Tenant Share Links

Status: tenant CRUD and owner enrollment are implemented in the tenant module;
frontend and public tenant links remain planned, 2026-11-11.

## Ownership and folders

Creating an organization is tenant creation. It belongs to the existing tenant
module together with list, get, update, and archive. Owner enrollment describes
the business journey; it is not a separate domain or deployable service.

```text
User FE -> User Gateway handler/tenant -> IdentityTenantService
        -> Identity domain/tenant -> repository/tenant -> Identity PostgreSQL
```

| Owner | Files |
| --- | --- |
| HTTP CRUD | `be/gateways/user/transport/http/handler/tenant/` |
| HTTP create request | `be/gateways/user/transport/http/request/tenant.go` |
| Identity errors and mappings | `be/services/identity/internal/exception/tenant*.go` |
| HTTP errors and localization | `be/gateways/user/transport/exception/` |
| Domain rules and interfaces | `be/services/identity/internal/domain/tenant/` |
| Atomic persistence and replay | `be/services/identity/internal/repository/postgres/tenant/create.go` |
| SQL source | `be/services/identity/internal/storage/postgres/query/tenant.sql` |
| Private gRPC | `be/services/identity/internal/transport/grpc/tenant/` |
| Source contract | `contracts/proto/identity/v1/tenant.proto` |

The Gateway router registers tenant CRUD independently of Catalog. Catalog
continues to own court/image behavior and its existing venue composition.

## Tenant CRUD contract

| Method | Route | Behavior |
| --- | --- | --- |
| POST | `/api/v1/tenant` | Create a pending tenant and assign its owner atomically |
| GET | `/api/v1/tenant` | List organizations visible through active membership |
| GET | `/api/v1/tenant/{organization_id}` | Read an organization in the caller's scope |
| PATCH | `/api/v1/tenant/{organization_id}` | Owner updates permitted organization fields |
| DELETE | `/api/v1/tenant/{organization_id}` | Owner archives a pending organization; 204 response |

Existing `/api/v1/tenant/organizations` CRUD URLs remain compatible. Their POST
is owner-only and retains its original organization-only response. All creation
runs through `IdentityTenantService.CreateOrganization` and one tenant domain
use case. Legacy RPC requests without a key require an existing owner and get
a per-call key; they do not gain retry replay guarantees.

The detailed [Tenant Creation API Contract](08-owner-onboarding-api-contract.md)
and [User OpenAPI](../../contracts/openapi/user-gateway.yaml) define the
canonical POST fields, errors, retry behavior, and response.

## Creation rules

- Derive the actor from a verified access token. Never accept actor IDs, roles,
  status, membership, or admin flags from the body.
- Accept active customers and existing owners. Reject suspended/admin accounts
  and staff roles; staff-to-owner support needs an explicit multi-role decision.
- Accept `name` and `slug`, plus a UUID v4 `Idempotency-Key` header. Validate
  the complete input once, then trim the name and trim/lowercase the slug.
- Retain name/slug limits of 120/100 characters and the slug pattern
  `^[a-z0-9]+(?:-[a-z0-9]+)*$`.
- Lock and recheck the current account, update its owner role, create a
  `PENDING_VERIFICATION` organization and `ACTIVE` owner membership, and store
  the replay result in one transaction.
- Scope replay by actor, operation, key, and normalized payload. A matching
  request replays the committed body for 24 hours; a changed payload or taken
  slug produces a distinct conflict. Recheck eligibility before replay.
- Retain the existing `onboard_requests` table and migration 000004. The module
  consolidation needs no schema rename or deletion of stored replay results.

Identity validates direct gRPC calls as well as HTTP requests. PostgreSQL owns
uniqueness and concurrency guarantees. Structured gRPC errors use the tenant
error domain, safe reason codes, and field violations; Gateway maps them into
localized English/Vietnamese responses without exposing SQL or credentials.

## Frontend and public tenant links

| Portal | Planned route | Behavior |
| --- | --- | --- |
| User | `/become-owner` | Authenticated customer submits tenant name and slug |
| User | `/t/:tenantSlug` | Public active tenant profile and published venues |
| User | `/t/:tenantSlug/venues/:venueSlug` | Public published venue within that tenant |
| User | `/login?next=...`, `/register?next=...` | Preserve a validated local destination |
| Tenant | Existing dashboard | Show organization status and copy its customer link |

The confirmed customer URL uses the organization slug:
`https://<user-domain>/t/smash-center`. Venue slugs identify venues within that
organization. Identity resolves tenant/venue scope and visibility; Catalog owns
courts and images. Services never join each other's databases.

Customer accounts remain platform accounts. Following a tenant link does not
create staff membership. Pending/inactive organizations and unpublished venues
remain hidden, with consistent not-found results for unavailable public data.
An active tenant with no published venues has an empty state.

After creation, the FE refreshes user/tenant state, shows pending review, and
links to Tenant portal login. Portal origins and token namespaces remain
separate; tokens never travel in URLs. Approval, venue publication, staff
invitations, public reads, and FE remain separate tracked work.

## Verification and rollout

Run the [isolated backend checks](../../be/tests/integration/identity/README.md)
using `bash be/tests/e2e/identity/tenant.sh`. Verify creation, CRUD, tenant scope,
legacy routes, replay/conflicts, rollback, concurrency, and account restrictions.
Keep local backend evidence separate from FE/deployment evidence.

Apply migration 000004 before Identity, then deploy the Gateway that uses the
additive tenant RPC fields. Roll back application images while retaining the
schema, roles, and replay history; do not automatically run migration down.
