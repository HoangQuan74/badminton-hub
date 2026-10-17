# Backend role

Own behavior inside the service boundary and keep the contract explicit.

- Trace contract -> handler/transport -> domain service -> repository -> registered route before saying an endpoint exists.
- Keep PostgreSQL authoritative. Use the owning service schema and migration directory; never query another service's schema directly.
- Protect booking correctness with database constraints and transactions. Redis may optimize, but cannot be the only overlap or idempotency boundary.
- Make retries and callbacks idempotent. Record stable operation or event identities where the workflow requires them.
- Keep prices server-calculated and snapshot the value used by a booking or payment.
- Regenerate generated SQL or contract artifacts after source changes and run the repository's explicit drift checks.
- Use sqlc for PostgreSQL access. Put cross-package integration, contract, and E2E tests in `be/tests`; keep package-local unit tests beside source only when white-box access is necessary.
- Cover authorization, invalid input, duplicate requests, concurrency, timeout, downstream failure, and rollback behavior.

Useful source areas: `be/services`, `contracts`, `docs/05-api-and-events.md`, `docs/08-backend-engineering-practices.md`.
