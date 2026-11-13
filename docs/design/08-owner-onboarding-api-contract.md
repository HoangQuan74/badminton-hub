# Tenant Creation API Contract

Status: backend implemented and verified locally for IDENTITY-01, 2026-11-11.
Frontend and deployment verification remain outstanding. The authoritative HTTP schema/examples live in
[User Gateway OpenAPI](../../contracts/openapi/user-gateway.yaml). This document
explains caller behavior and the transaction, not runtime availability.

Related: [Folder and Flow Design](07-owner-onboarding-and-tenant-links.md) and
[Identity Tasks](../tasks/identity.md).

## Request

```http
POST /api/v1/tenant
Authorization: Bearer <existing-access-token>
Content-Type: application/json
Idempotency-Key: 84f6cb1a-27b1-4e9b-9a14-95e36cfbbac6
Accept-Language: vi
```

```json
{
  "name": "Smash Center",
  "slug": "smash-center"
}
```

- Existing customer authentication is enough; the caller does not need an
  owner role or organization membership before enrolling.
- Accept active customers and existing owners. Existing owners may create
  another organization with another slug. Customer capabilities remain usable.
- Reject invalid/missing/deleted-account tokens with `401`; suspended/admin
  actors with `403`; existing staff roles with `409 TENANT_ROLE_CONFLICT`.
  Staff-to-owner support requires the separate multi-role decision.
- Require one valid UUID v4 key, JSON content type, and at most 4096 encoded body
  bytes. Reject unknown properties; no user ID, email, password, role, membership,
  approval state, or admin flag belongs in this body.
- Name: required, at most 120 Unicode characters raw and trimmed; trim outer
  whitespace and reject an empty result or a NUL character. Preserve inner spacing and case.
- Slug: required, at most 100 characters raw and normalized; trim/lowercase and
  require `^[a-z0-9]+(?:-[a-z0-9]+)*$`. Uppercase input is canonicalized. The
  database enforces uniqueness across pending/active/rejected/suspended/archived
  organizations. No reserved-name policy is introduced here.
- Optional `?lang=vi|en` overrides `Accept-Language`; English is the fallback.
  Invalid query-language values return `400` with the fallback message language.

## Success

Return `201 Created` after committing all writes. The response has three fields:

| Field | Contents |
| --- | --- |
| `user` | Existing public User shape with `roles: [tenant_owner]`, active status, and `is_admin: false` |
| `organization` | ID, normalized name/slug, `PENDING_VERIFICATION`, and UTC created/updated timestamps |
| `membership` | Organization ID, caller user ID, role `tenant_owner`, status `ACTIVE` |

The full success example is in OpenAPI. Return no access/refresh tokens or
password/hash material. Existing JWTs identify the same user; subsequent user
lookup resolves the updated role. Do not create a second login session.

Headers: `Location: /api/v1/tenant/{id}`,
`Idempotency-Replayed: false|true`, `Idempotency-Key-Expires-At` (UTC deadline),
`X-Request-ID` for the current request, and `Cache-Control: no-store`.
Errors also return `X-Request-ID` and `Cache-Control: no-store`.

The organization is still awaiting admin review. Active membership permits the
owner to read/manage its pending application, but does not activate the
organization, create a venue, or expose the public tenant link. IDENTITY-02/03
remain the approval/publication boundary.

## Transaction and idempotency

1. Authenticate the current actor and validate/canonicalize input.
2. Lock/recheck the account and enrollment eligibility inside the transaction.
3. Look up the key scoped by actor + operation before checking new slug creation.
   Compare a hash of normalized name/slug, excluding auth/language/request IDs.
4. Replay an existing matching committed result, or reject a mismatched payload.
5. For a new attempt, persist the fixed owner role, pending organization, active
   owner membership, and response snapshot atomically. Roll back all on failure.

Proposed replay retention is **24 hours from the original commit**, disclosed by
`Idempotency-Key-Expires-At`. Replay never extends it. Same payload/key returns
the original body and `201`, with `Idempotency-Replayed: true`. Account eligibility
is checked again, so a suspended account cannot replay a previous success.
The body is a historical snapshot; refetch the organization and `/me` for current
state. Request ID headers reflect the current request, not the saved attempt.

Concurrent matching requests produce one organization and replay its result.
Different-key requests competing for the same slug produce one success and one
slug conflict, with no partial role changes for the losing request. In-flight
matching requests wait within their deadline; a deadline does not imply the
original write failed. Requests failing before commit do not consume the key.

For uncertain network/500/502/504 outcomes, retry the same key and payload.
Do not issue a new key automatically. Within retention, a changed payload under
the old key returns `409`; an intentional corrected/new submission uses a new
key after resolving the previous outcome. After retention expires, inspect
current organizations before a new attempt; exactly-once replay is no longer
guaranteed. The UI retains unresolved attempts across reloads without storing
tokens in URLs.

## Error contract

Use the existing top-level `error` shape with required `code`, safe localized
`message`, and `request_id`. Optional `fields` entries contain `field`, `code`,
and `message`, never submitted values. Field `idempotency_key` refers to the
`Idempotency-Key` header. Malformed JSON may have no field details.

| HTTP | Public code | Caller action |
| --- | --- | --- |
| 400 | `INVALID_REQUEST` | Correct indicated fields/headers; unknown properties are invalid |
| 401 | `UNAUTHORIZED` | Existing refresh/login flow; retain the attempt and destination |
| 403 | `FORBIDDEN` | Explain denied enrollment; no automatic retry |
| 409 | `ORGANIZATION_SLUG_ALREADY_EXISTS` | Show slug field error; resolve any earlier attempt before changing key |
| 409 | `IDEMPOTENCY_KEY_CONFLICT` | Resolve original submission; do not blindly resubmit |
| 409 | `TENANT_ROLE_CONFLICT` | Explain staff-role restriction; existing access remains unchanged |
| 413 | `PAYLOAD_TOO_LARGE` | Reduce payload; no writes |
| 415 | `UNSUPPORTED_MEDIA_TYPE` | Send application/json; no writes |
| 500 | `INTERNAL_ERROR` | Show generic error/request ID; retain key for uncertain outcome |
| 502 | `DEPENDENCY_UNAVAILABLE` | Bounded retry with same key/payload |
| 504 | `DEPENDENCY_TIMEOUT` | Outcome unknown; retry same key/payload |

gRPC domain failures use standard statuses plus `google.rpc.ErrorInfo` for the
stable reason and `BadRequest` for field violations. Slug conflict and key
conflict can share `AlreadyExists` but must have different reasons. HTTP-only
body-size/content-type checks remain in Gateway. Unknown details fall back to a
safe status mapping. Never parse message strings or return raw SQL/gRPC errors.

Example field error:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Vui lòng kiểm tra thông tin đã nhập.",
    "request_id": "req_example",
    "fields": [
      {
        "field": "slug",
        "code": "SLUG_INVALID",
        "message": "Slug chỉ gồm chữ cái, chữ số và dấu gạch nối giữa các phần."
      }
    ]
  }
}
```

## Tenant CRUD and compatibility

All tenant HTTP CRUD adapters live under `handler/tenant`. The canonical routes
are `POST /api/v1/tenant`, `GET /api/v1/tenant`, and
`GET/PATCH/DELETE /api/v1/tenant/{organization_id}`. Read/update/archive keep
existing membership, owner authorization, and lifecycle constraints; DELETE
archives only pending organizations and returns 204.

Existing `/api/v1/tenant/organizations` CRUD URLs remain compatibility routes.
Their POST retains the original owner-only request (`name`, `slug`) and
organization-only response. Internal callers without an idempotency key remain
owner-only; Identity generates a per-call key and executes the same atomic
create use case. These legacy calls do not provide retry replay guarantees.
New callers use the canonical POST and supply `Idempotency-Key`.

The separate unpublished onboarding HTTP/RPC interface has been removed.
The existing `onboard_requests` table and migration 000004 are retained so no
schema rename or replay-history deletion is required by the folder refactor.

## Internal contract

[`identity/v1/tenant.proto`](../../contracts/proto/identity/v1/tenant.proto) defines `IdentityTenantService.CreateOrganization` with
request fields `access_token`, `name`, `slug`, and
`idempotency_key`. No actor ID is accepted. Request/correlation IDs and deadlines
are propagated as metadata/context. Response uses existing `UserResponse` and
`TenantOrganization` plus a typed owner membership and internal replay metadata
(`idempotency_replayed`, `idempotency_key_expires_at`) for HTTP response headers.
Replay metadata does not add fields to the public response body. This private RPC still
requires the Identity domain validation even if HTTP has already validated.

The Protobuf source/generated stubs, Identity handler, PostgreSQL adapter, and
User Gateway route are implemented. Structured errors/localization are scoped
to the new endpoint; unrelated Gateway error behavior is unchanged.

## FE integration and acceptance

- Enroll on the User origin after register/login. Keep one key and normalized
  payload per unresolved attempt; show validation/conflict errors inline and
  success as pending review. Refresh user/org state after success or replay.
- Compute the future share link as `/t/{organization.slug}` on the configured
  User origin; explain that pending tenants are not publicly available yet.
- User/Tenant portals retain separate sessions; provide the Tenant login link
  without passing tokens through redirects. No automatic cross-origin SSO.
- CORS must allow `Idempotency-Key` and expose `Location`,
  `Idempotency-Replayed`, `Idempotency-Key-Expires-At`, and request-ID headers.
- Verify direct gRPC validation, unknown JSON properties, raw/normalized limits,
  whitespace-only name, uppercase slug normalization, role denial, all three
  conflicts, concurrent key/slug requests, rollback, replay after suspension,
  language precedence, and timeout/reload recovery.
- Run contract example checks before implementation. Record PostgreSQL,
  Gateway, browser, and deployment evidence separately when implemented.

## Verification and rollout

Run [the isolated backend checks](../../be/tests/integration/identity/README.md)
with `bash be/tests/e2e/identity/tenant.sh`. This uses real HTTP adapters, a real
Identity process, and a separately migrated PostgreSQL database with no seeds.

Deploy migration `000004_create_onboard` first, Identity second, User Gateway
last. Previous Gateways remain compatible with the additive tenant RPC fields.
Application rollback changes images and retains the schema, system roles, and
replay records. Do not automatically run migration down: it removes replay
history, while created organizations/memberships remain.

Replay expiry controls eligibility, not immediate physical removal. An expired
key is removed when reused inside a new transaction; background pruning of
older records is an operational follow-up. Preserve unresolved keys in the FE.
