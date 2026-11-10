---
name: devops
description: Operate and improve BadmintonHub Docker, Compose, CI/CD, deployment, configuration, observability, and recovery workflows safely.
---

# BadmintonHub DevOps

Use this project role for local infrastructure, CI/CD, deployment, runtime diagnosis, observability, and recovery planning.

## Context loading

Before changing infrastructure, always read `docs/09-project-conventions.md` first.
Treat its ownership, configuration, folder, naming, test, and verification rules
as mandatory. Then read `docs/ai-context/INDEX.md`,
`docs/06-testing-and-devops.md`, `docs/07-delivery-plan.md`,
`docs/design/06-microservice-migration.md`, `devops/README.md`,
`devops/RELEASE.md`, the relevant Compose files, Dockerfiles, scripts, and
Jenkins pipeline.

## Working rules

- Resolve the environment, service, image, network, database, namespace, and owner before mutation.
- Keep service-scoped configuration: `IDENTITY_*`, `GATEWAY_*`, `CATALOG_*`, `BOOKING_*`, `PAYMENT_*`, and process-specific prefixes.
- Separate build success, deployment success, readiness, dependency health, runtime behavior, and business verification.
- Use health and readiness checks, bounded timeouts, structured logs, metrics, traces, and explicit rollback paths.
- Preserve data: do not reset volumes, drop schemas, force migrations, or retry business operations without authorization and idempotency evidence.
- Kafka, PostgreSQL, Redis, MongoDB, and OpenTelemetry Collector failures must be observable and recoverable.
- Report exact commands, environment limitations, evidence, and remaining unverified gates.

## Additional project checks

- Keep local Compose defaults separate from deployed configuration. Check migration compatibility, backups, immutable image tags, and rollback for delivery changes.
