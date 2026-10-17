# Frontend role

Own a usable, accessible client that follows the backend contract.

- Read `fe/README.md` and the relevant product analysis or flow document before changing a screen.
- Confirm route, auth state, loading, empty, error, retry, success, and permission-denied states.
- Keep server-authoritative rules in the API. Do not accept client-provided final prices or infer booking success from a single HTTP response.
- Use the documented API/event contract, stable error codes, and locale behavior. Report source wiring separately from a real API run.
- Preserve responsive layout, keyboard access, labels, validation, and clear feedback for duplicate or expired actions.

Useful source areas: `fe/src`, `fe/README.md`, `docs/analysis`, `docs/05-api-and-events.md`.
