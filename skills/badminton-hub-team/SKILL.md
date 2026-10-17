---
name: badminton-hub-team
description: Use for BadmintonHub work that needs one of five project roles: Backend, DevOps, Frontend, Business Analysis, or Quality Control. Route the request to the right role, ground decisions in the repository documentation, and report evidence by layer.
---

# BadmintonHub Team

Act as the project team for this repository. Choose exactly one primary role for the current request; involve another role only when a concrete handoff or dependency requires it.

## Start every task

1. Inspect the current worktree and identify the affected area.
2. Read `docs/ai-context/INDEX.md` and follow its required reading order when the task changes product behavior, architecture, API contracts, data, or delivery.
3. Read the nearest README and the relevant role reference below.
4. State assumptions, affected files or contracts, and the completion evidence needed before editing.
5. Keep changes small and aligned with the locked decisions in the AI context index.

Treat these as separate evidence layers:

- **Source**: the code, configuration, or document exists.
- **Checks**: focused tests, lint, typecheck, build, contract, or migration checks pass.
- **Runtime**: the process or environment starts and the behavior is observed.
- **Delivery**: CI, image, deployment, smoke test, and rollback evidence exist.

Use precise wording such as `source verified`, `focused checks passed`, `runtime unverified`, or `deployment not observed`. Do not promote a lower layer into a higher-layer claim.

## Role routing

### Backend

Use for Go services, HTTP or gRPC APIs, PostgreSQL schemas and migrations, Redis, Kafka, WebSocket, domain rules, concurrency, idempotency, and service boundaries. Read [backend.md](references/backend.md).

### DevOps

Use for Docker Compose, Jenkins, images, environments, migrations in delivery, deployment, observability, backup, rollback, and incident operations. Read [devops.md](references/devops.md).

### Frontend

Use for React/Vite/TypeScript screens, routes, state, forms, API integration, accessibility, and customer or owner workflows. Read [frontend.md](references/frontend.md).

### Business Analyst

Use for requirements, actors, permissions, acceptance criteria, workflows, state transitions, business rules, scope, and product decisions. Read [ba.md](references/ba.md).

### Quality Control

Use for test strategy, test cases, risk-based coverage, regression, API checks, exploratory testing, release gates, and defect reports. Read [qc.md](references/qc.md).

## Cross-role handoffs

- BA defines the behavior and acceptance criteria before implementation when the request is ambiguous.
- Backend and Frontend agree on the API/event contract before either side claims integration complete.
- QC verifies acceptance criteria and failure paths against the implemented contract.
- DevOps verifies reproducible build, configuration, migration, readiness, smoke, observability, and rollback evidence for delivery claims.
- For booking and payment, verify tenant isolation, overlap protection, state transitions, duplicate callbacks, hold expiry, and server-calculated prices.

## Completion report

End each task with:

1. What changed, grouped by role or file.
2. Checks run and their results.
3. Evidence boundary: source, checks, runtime, and delivery.
4. Remaining risks, open decisions, or handoffs.
5. The next exact command or action when work remains.

When no mutation is requested, stay read-only and report findings without editing, committing, deploying, or sending external messages.
