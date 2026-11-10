---
name: qc
description: Plan and execute BadmintonHub quality checks across contracts, Go services, React frontend, integrations, failure modes, and release evidence.
---

# BadmintonHub Quality

Use this project role for test planning, regression testing, release gates, bug reproduction, and evidence-based quality review.

## Context loading

Before planning or executing checks, always read `docs/09-project-conventions.md`
first. Treat its ownership, contract, test-location, formatting, and evidence
rules as mandatory. Then read `docs/ai-context/INDEX.md`,
`docs/06-testing-and-devops.md`, `docs/05-api-and-events.md`,
`docs/08-backend-engineering-practices.md`, the relevant requirements and
state-machine documents, changed source, and deployment/configuration path.

## Test model

- Verify contracts first: route or event shape, auth, permissions, status codes, errors, idempotency, and compatibility.
- Cover unit behavior, integration behavior with real infrastructure when correctness depends on it, contract checks, and critical end-to-end journeys.
- For booking, payment, and Kafka flows, test duplicates, retries, concurrency, timeouts, partial failure, outbox recovery, consumer idempotency, and durable final state.
- For frontend work, test auth bootstrap, route guards, API error states, permission states, responsive and accessibility behavior, and build output.
- Treat `TEST_DATABASE_URL` skips, Docker or socket `EPERM`, missing VPN or credentials, unavailable dependencies, and timeouts as evidence limits rather than passes.
- Separate code and test results from deployment, runtime, provider, settlement, and user-acceptance evidence.
- Report failures with exact reproduction, expected versus actual behavior, evidence, severity, and the smallest next action.

## Completion

A check is complete only when it has a reproducible command or artifact, a clear pass/fail result, and an explicit evidence boundary.

## Additional project checks

- Prioritize tenant isolation, double-booking races, duplicate payment callbacks, hold expiry, auth refresh/logout, localized errors, and migration compatibility when relevant.
