# Repository Guidelines

## Project Structure & Module Organization

- `be/services/` contains Go services; `be/gateways/` contains HTTP and WebSocket gateways. Each service and gateway is an independent Go module. Shared backend code lives in `be/platform/`.
- `fe/src/` contains the React/TypeScript app, grouped by feature, with shared components, routing, and styles. Frontend test material is under `fe/tests/`.
- `contracts/` holds Protobuf, OpenAPI, event definitions, and generated code. `devops/` contains Compose, deployment, and monitoring files. Start with `docs/README.md` and `docs/ai-context/INDEX.md` for project context.

## Build, Test, and Development Commands

Run from the repository root unless noted:

- `make -C be test` runs Go tests across backend modules; `make -C be vet` runs Go vet.
- `make -C be build` builds the three gateways and Identity. `make -C be coverage-platform` checks shared platform coverage.
- `cd fe && npm ci` installs frontend dependencies; `npm run dev:all` starts the local portals. Use `npm run lint`, `npm run typecheck`, and `npm run build` to validate frontend changes.
- For the local Compose stack, environment setup, migrations, and service commands, follow `devops/README.md`.

## Coding Style & Naming

Format Go code with `gofmt` and use standard Go naming and tab indentation. Follow existing TypeScript conventions: two-space indentation, PascalCase components, and feature-oriented files such as `auth.api.ts`. Run ESLint for frontend changes. Edit source contracts and SQL queries rather than generated files; use `make -C be generate` for Identity sqlc output.

## Testing Guidelines

Name Go tests `*_test.go` and use `Test...` functions; prefer table-driven tests for multiple cases. Run focused tests within the affected module with `go test ./...`, then run the relevant broader checks. The `be/platform` module has a 100% coverage gate. For integration scenarios, consult `docs/06-testing-and-devops.md` and use real dependencies where appropriate. Frontend currently has no test script; run lint, typecheck, and build.

## Commits & Pull Requests

Recent commit subjects use short prefixes such as `feat:`, `fix:`, and `ref:`. Keep subjects concise and describe the change. In pull requests, explain the behavior and scope, link relevant issues, and list checks run; include screenshots for visible UI changes. Call out migrations, configuration changes, and verification that remains outstanding.

## Security & Configuration

Use the provided `.env.example` files for local configuration and keep secrets out of commits and logs. Replace local-only credentials and JWT secrets for deployed environments. Preserve service-owned database boundaries and tenant isolation.
