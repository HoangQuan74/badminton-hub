# Identity Tasks

Track the remaining Identity work identified by source review on 2026-11-11.
All tasks below are `TODO`; this document records planned work, not completed
implementation or runtime verification. Priorities and sequencing are proposed.

## Scope and current baseline

Identity owns users, credentials, sessions, token issuance, roles, permissions,
organizations, venues, organization membership, and venue staff assignments.
Public HTTP routes belong to User/Admin Gateway; internal calls use versioned
gRPC contracts. Catalog continues to own courts and venue images.

Source already includes User/Tenant/Admin authentication, User/Staff/Role CRUD,
permission assignment and listing, organization/venue creation and updates, and
staff assignment. These capabilities still need the completion work below.
Do not treat older migration notes saying Admin Auth/Role is unimplemented as
the current baseline.

Requirements: [Owner Onboarding](../analysis/01-owner-onboarding.md),
[Venue Setup](../analysis/02-venue-setup.md),
[Product Requirements](../01-product-requirements.md),
[API and Events](../05-api-and-events.md), and
[Migration Security](../design/06-microservice-migration.md#security).
Implementation reference: [Identity README](../../be/services/identity/README.md).

## Task register

`P0` closes an existing business-flow or authorization gap. `P1` completes
administration, account management, or production readiness. `MVP` follows
existing requirements; `Follow-up` is a proposed extension; `Production` is
hardening to finish before deployment beyond the local trust model.

| ID | Task | Priority | Scope | Status | Depends on |
| --- | --- | --- | --- | --- | --- |
| IDENTITY-01 | Complete owner onboarding | P0 | MVP | TODO | IDENTITY-04 |
| IDENTITY-02 | Review organizations and audit decisions | P0 | MVP | TODO | IDENTITY-04 |
| IDENTITY-03 | Submit, approve, and publish venues | P0 | MVP | TODO | IDENTITY-02 |
| IDENTITY-04 | Enforce permissions and active-account checks | P0 | MVP | TODO | None |
| IDENTITY-05 | Invite staff and list memberships/assignments | P1 | MVP | TODO | IDENTITY-01, IDENTITY-04 |
| IDENTITY-06 | Add self-service profile updates | P1 | Follow-up | TODO | None |
| IDENTITY-07 | Complete password change and recovery | P1 | Follow-up | TODO | IDENTITY-09 |
| IDENTITY-08 | Expose User administration through Admin Gateway | P1 | MVP | TODO | IDENTITY-04 |
| IDENTITY-09 | Complete session management and revocation | P1 | Production | TODO | IDENTITY-04 |
| IDENTITY-10 | Protect authentication against repeated attempts | P1 | Production | TODO | None |
| IDENTITY-11 | Authenticate internal gRPC callers | P1 | Production | TODO | None |
| IDENTITY-12 | Migrate JWT signing to asymmetric keys and JWKS | P1 | Production | TODO | None |
| IDENTITY-13 | Make health checks reflect dependency failures | P1 | Production | TODO | None |
| IDENTITY-14 | Complete verification evidence and refresh documentation | P1 | MVP + Production | TODO | Applicable release tasks |

PIC and target dates are unassigned. Update the register and task checkboxes
together. Mark a task `Done` only when its acceptance checks have evidence;
record skipped checks and missing runtime/deployment evidence separately.

## Phase 1: owner onboarding and authorization

### IDENTITY-04 — Enforce permissions and active-account checks

- [ ] Require an active account for protected User/Tenant/Admin operations;
  reject an already-issued token after the account is suspended.
- [ ] Define and enforce an operation-to-permission matrix in Identity. Document
  the platform administrator bypass and how existing admin accounts migrate.
- [ ] Load assigned permissions in Role/Staff/current-staff responses. Currently
  role-permission writes exist, but role reads do not populate those grants.
- [ ] Prevent unauthorized privilege changes; audit role, permission, and access
  mutations without recording credentials.
- **Acceptance/tests:** verify allowed and denied actions through gRPC and
  Gateway; cover suspended actors, missing grants, self-escalation, permission
  replacement rollback, and preservation of tenant/venue boundaries.

### IDENTITY-01 — Complete owner onboarding

- [ ] Define an explicit owner registration or account-upgrade use case. Normal
  registration must not accept arbitrary client-supplied roles or admin flags.
- [ ] Remove the current dead end: registration creates a regular user, while
  organization creation requires the global `tenant_owner` role.
- [ ] Provision required system roles in versioned migrations rather than
  relying on local development seeds. Preserve existing customer access.
- [ ] Persist the required role/membership changes atomically and define safe
  behavior for repeated or concurrent onboarding requests.
- **Acceptance/tests:** a newly registered account can follow the chosen owner
  flow, create its first organization in `PENDING_VERIFICATION`, and read only
  its own organization without manual SQL or seeded owner credentials.

### IDENTITY-02 — Review organizations and audit decisions

- [ ] Add authorized Admin Gateway and Identity gRPC operations to list/filter
  organizations, inspect details, approve, reject, and suspend them.
- [ ] Define allowed transitions from existing organization states, including
  repeat/conflicting decisions and whether reinstatement is supported.
- [ ] Commit each decision and its actor, previous/new state, reason, timestamp,
  and request ID in an Identity-owned audit record in the same transaction.
- [ ] Define the access matrix for pending, rejected, and suspended organizations;
  preserve the rule that only active organizations publish venues.
- **Acceptance/tests:** approve a pending organization and create a venue through
  the Gateway; cover missing permission, invalid transitions, concurrent admin
  decisions, audit rollback, and suspension effects on public visibility/access.

### IDENTITY-03 — Submit, approve, and publish venues

- [ ] Add owner submission and admin review/publication/suspension operations.
  The current creation path stops at `DRAFT`.
- [ ] Define the transition matrix around `DRAFT`, `PENDING_APPROVAL`,
  `PUBLISHED`, `SUSPENDED`, and `ARCHIVED`. Decide rejection/resubmission behavior
  without silently adding a new state to the existing contract.
- [ ] Define publication prerequisites with Catalog's court/image ownership and
  the planned operating-hours/pricing owners. Keep missing prerequisites visible;
  do not query another service's schema or treat missing features as complete.
- [ ] Enforce active organization and authorized actor checks; persist review
  decisions with audit records and safe concurrent transition handling.
- **Acceptance/tests:** a submitted, approved venue appears in the marketplace;
  draft, unapproved, archived, and suspended venues remain hidden. Cover wrong
  tenant, missing prerequisites, duplicate approval, and organization suspension.

## Phase 2: staff and account management

### IDENTITY-05 — Invite staff and list memberships/assignments

- [ ] Add paginated membership and venue-assignment reads for authorized owners.
  Keep existing tenant-wide versus assigned-venue access semantics.
- [ ] Define invitation creation, acceptance, expiry, cancellation, and duplicate
  handling for new/existing accounts; derive the target tenant from authorized
  owner scope and restrict which roles may be granted.
- [ ] Close the current limitation that assignment requires an existing active
  user with a matching global role. Preserve customer capabilities.
- [ ] Separate invitation state from delivery; email delivery may be deferred
  under the delivery plan, but acceptance must be demonstrable without SQL.
- **Acceptance/tests:** invited staff can accept once and access only granted
  scope. Cover expired/replayed invitations, wrong recipient, cross-tenant
  assignment, membership deactivation, and immediate loss of revoked access.

### IDENTITY-06 — Add self-service profile updates

- [ ] Add User/Tenant profile updates for agreed editable fields, starting with
  name and phone; derive the target user from the verified token.
- [ ] Keep role, status, membership, and admin flags outside self-service input.
  Decide email-change verification separately before allowing email edits.
- **Acceptance/tests:** update the caller's profile through the Gateway; cover
  invalid fields, inactive accounts, another user's ID, and privilege fields.

### IDENTITY-07 — Complete password change and recovery

- [ ] Add authenticated password change for User/Tenant and align validation
  across all actors at the Identity boundary, including direct gRPC calls.
- [ ] Design reset request/confirmation with expiring, single-use tokens stored
  as hashes, enumeration-safe responses, and an explicit delivery mechanism.
- [ ] Change the password and revoke affected sessions atomically using
  IDENTITY-09. Define whether the current session is retained or replaced.
- **Acceptance/tests:** cover wrong current password, password validation, expired
  or reused reset tokens, concurrent reset attempts, and old-session rejection.

### IDENTITY-08 — Expose User administration through Admin Gateway

- [ ] Connect existing Identity User CRUD gRPC operations to authorized Admin
  Gateway routes and publish the public contracts.
- [ ] Support existing pagination/search/date filters and enforce per-operation
  permissions, including account suspension.
- [ ] Review deletion versus existing membership/history constraints; document
  rejection or suspension behavior instead of bypassing foreign keys.
- **Acceptance/tests:** exercise list/detail/create/update/suspend and the agreed
  deletion policy; cover unauthorized actors, duplicate email, malformed IDs,
  suspended-account access, and membership-linked users.

## Phase 3: authentication and operational readiness

### IDENTITY-09 — Complete session management and revocation

- [ ] Add caller-scoped session listing, revoke-one, and revoke-all operations;
  never return refresh tokens or their stored hashes.
- [ ] Integrate password changes and account suspension with session revocation.
  Define access-token invalidation guarantees, including Socket consumers.
- [ ] Review refresh rotation across portal boundaries: validate eligibility
  before committing a replacement session; preserve atomic single-use rotation.
- [ ] Decide refresh replay/family handling and expired-session cleanup policy.
- **Acceptance/tests:** use PostgreSQL to verify concurrent refresh has one
  winner, wrong-portal refresh does not consume a valid session, revoked sessions
  cannot refresh, and one user cannot list or revoke another user's sessions.

### IDENTITY-10 — Protect authentication against repeated attempts

- [ ] Define bounded account/source limits for login, refresh, and recovery;
  avoid an account-lockout policy that lets another caller lock out a victim.
- [ ] Enforce limits across replicas with explicit dependency-failure behavior;
  prevent direct gRPC access from bypassing the intended protection.
- [ ] Document stable throttling responses, recovery behavior, and metrics;
  exclude passwords and tokens from logs and metric labels.
- **Acceptance/tests:** cover repeated failures, expiry/recovery, parallel
  attempts, account enumeration, and limiter dependency failure.

### IDENTITY-11 — Authenticate internal gRPC callers

- [ ] Choose and document service authentication and transport protection;
  validate caller identity rather than trusting `x-service-identity` metadata.
- [ ] Restrict each caller to allowed RPCs while independently validating user
  tokens and tenant scope. Keep authorization rules outside platform mechanics.
- [ ] Document credential provisioning, rotation, health-probe access, staged
  enforcement, and rollout/rollback across Identity, Gateways, and Catalog.
- **Acceptance/tests:** reject missing/invalid credentials, spoofed service
  metadata, and disallowed callers; verify valid calls retain deadlines and IDs.

### IDENTITY-12 — Migrate JWT signing to asymmetric keys and JWKS

- [ ] Keep private signing keys in Identity and distribute public verification
  keys through a versioned JWKS contract, as required by the migration design.
- [ ] Define algorithm, issuer, audience, key IDs, verifier caching, and rotation;
  update every verifier, including Gateways, Catalog, and Socket as applicable.
- [ ] Plan a bounded HS256 compatibility window for existing sessions and a
  rollback that does not distribute private keys to verifiers.
- **Acceptance/tests:** cover valid/expired tokens, wrong issuer/audience,
  unknown or retired keys, key rotation, JWKS outage, and all verifier consumers.

### IDENTITY-13 — Make health checks reflect dependency failures

- [ ] Make Identity readiness reflect PostgreSQL availability after startup,
  rather than leaving gRPC health permanently `SERVING` until shutdown.
- [ ] Make User/Admin Gateway readiness follow required upstream health with
  bounded probes; keep process liveness independent from dependency readiness.
- [ ] Replace TCP-only checks where appropriate and document failure/recovery
  signals in Compose/release configuration and the operational runbook.
- **Acceptance/tests:** stop/recover a disposable DB or Identity instance;
  readiness fails then recovers, liveness remains meaningful, and graceful
  shutdown drains requests within the configured timeout.

## Verification and completion

### IDENTITY-14 — Complete verification evidence and refresh documentation

- [ ] Add focused unit and PostgreSQL integration checks for each implemented
  task. Existing DB tests skip without `TEST_IDENTITY_DATABASE_URL`; record that
  outcome as unverified, and require an isolated migrated database in CI.
- [ ] Add contract coverage and a Gateway E2E journey: fresh registration →
  owner onboarding → organization approval → venue creation/submission/approval
  → public visibility → staff assignment → access revocation.
- [ ] Regenerate SQLC/Protobuf artifacts when their sources change and run the
  relevant drift checks. Keep local seeds out of Dev/Production rollout.
- [ ] Update OpenAPI, service README, AI context, API/event reference, and stale
  migration notes to match delivered behavior; preserve legacy gRPC endpoints
  until the compatibility removal gate is met.
- [ ] Record schema/config changes, deployment order, backward compatibility,
  migration verification, image rollback, and outstanding product decisions.

Run focused checks in affected modules first, then the applicable broader checks:

```bash
cd be/services/identity && go test ./...
```

From the repository root:

```bash
make -C be test
make -C be vet
make -C be build
```

Record source, automated-check, runtime, deployment, and business evidence
separately. A successful health response or skipped DB test does not establish
completion of an onboarding or authorization flow.

## Implementation decisions to resolve

- IDENTITY-01/05: explicit owner enrollment and the relationship between global
  portal roles and organization membership; prevent role-management dead ends.
- IDENTITY-02/03: suspension access, reinstatement, venue rejection/resubmission,
  and publication prerequisites owned by other service slices.
- IDENTITY-04: permission naming, administrator bypass, audit access, and safe
  provisioning for existing deployments without development seeds.
- IDENTITY-07/09: recovery delivery, session retention after password change,
  refresh replay policy, and the required access-revocation latency.
- IDENTITY-11/12: service credentials, key storage, rotation, compatibility
  windows, and the order in which consumers adopt new verification rules.

Resolve these in the owning task and update the nearest contract or ADR before
implementation depends on them; they are not accepted product decisions yet.
