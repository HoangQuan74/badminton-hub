---
name: fe
description: Build, review, and verify the BadmintonHub React frontend against API contracts, auth flow, UX states, accessibility, and deployment conventions.
---

# BadmintonHub Frontend

Use this project role for work under `fe/` and frontend-to-backend integration.

## Context loading

Before changing UI behavior, always read `docs/09-project-conventions.md` first.
Treat its ownership, naming, contract, test-location, and verification rules as
mandatory. Then read `docs/ai-context/INDEX.md`,
`docs/01-product-requirements.md`, `docs/05-api-and-events.md`,
`docs/06-testing-and-devops.md`, `fe/README.md`, `fe/package.json`, the router
and auth modules, and the relevant API contract.

## Working rules

- Verify method, path, auth, payload, response, error, permission, and state transition. Distinguish contract and source evidence from live API evidence.
- Preserve the existing React/Vite conventions, auth storage/bootstrap, route guards, query client, and `VITE_API_URL` behavior.
- Model loading, success, empty, validation, unauthorized, forbidden, conflict, dependency failure, retry, and stale-data states.
- Never put provider credentials, JWT signing secrets, or server-only configuration in the browser.
- Keep user, organization, venue, booking, payment, and role scope explicit in the UI.
- Verify keyboard access, labels, focus, responsive behavior, localization, and confirmation for destructive actions.
- Run focused lint, typecheck, and build checks. Report source verification separately from live API acceptance.

## Code quality principles

Write frontend code for a beginner maintainer as well as the current feature:

- Prefer clear, explicit code over clever or compressed expressions.
- Keep each component and hook focused on one responsibility; split a component when its data flow or rendering becomes difficult to scan.
- Use feature-based folders and names that describe the business responsibility, such as `features/tenant`, `features/admin`, and `features/bookings`.
- Keep route definitions, page composition, reusable UI, API calls, and state management in their own obvious layers.
- Prefer existing React, TypeScript, React Router, TanStack Query, and form patterns in this repository before introducing abstractions or dependencies.
- Keep props and API types explicit. Treat optional fields and permission boundaries deliberately instead of hiding them behind `any`, unchecked casts, or implicit defaults.
- Make the happy path and failure path readable from the same file: loading, empty, error, forbidden, retry, and success states should have named branches or components.
- Use small, reviewable changes. Avoid unrelated refactors, broad CSS overrides, and abstractions that do not have a second real consumer.
- Comment intent and business reasoning only where the code cannot explain it clearly; do not comment obvious syntax.
- Before considering a frontend change complete, verify that a new contributor can find the route, page, data source, permission rule, and validation command without reconstructing hidden conventions.
- Keep user-facing copy in the shared i18n message catalog; components consume translation keys instead of embedding Vietnamese or English strings.
- Support Vietnamese and English together from the first implementation: persist the selected language, update the document language, expose an accessible language switcher, and translate validation, loading, empty, error, permission, and action states.
- Add new translation keys in both supported languages before using them in a component. Keep proper names, venue names, and server-provided content separate from UI copy.
- Treat User, Tenant, and Admin as separate portal processes with separate origins and login entry points. Share components and utilities through the package, while keeping each process's root page, token namespace, API base URL, and auth path explicit.
- Use `customer` for the marketplace account, `tenant_owner` or `tenant_staff` for tenant operators, and `admin` for platform administration. Do not use a client-side route switch as a substitute for backend portal authorization.

## Additional project checks

- Keep final pricing and booking outcomes server-authoritative; verify the durable result when the workflow requires it.
