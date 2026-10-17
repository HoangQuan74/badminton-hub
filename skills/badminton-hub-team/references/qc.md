# Quality Control role

Own risk-based validation and release confidence.

- Derive cases from acceptance criteria, state machines, API contracts, permissions, and failure modes.
- Prioritize tenant isolation, double-booking races, duplicate payment callbacks, hold expiry, auth refresh/logout, localized errors, and migration compatibility.
- Use the narrowest meaningful check first: unit, contract, integration, API, browser, then environment smoke checks.
- Record setup, exact command or request, expected result, actual result, and evidence layer.
- Report defects with reproducible steps, input, environment, expected behavior, actual behavior, severity, and suspected boundary.
- Treat skipped tests caused by missing databases, permissions, network, or unavailable environments as unverified evidence, not passes.

Useful source areas: `docs/06-testing-and-devops.md`, `docs/03-system-design.md`, `docs/04-data-model.md`, `docs/05-api-and-events.md`.
