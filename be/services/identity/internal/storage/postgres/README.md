# Identity PostgreSQL storage

This directory owns the Identity persistence boundary:

- `schema/migrations/`: versioned schema changes for the `identity` schema.
- `schema/seed/`: local development data only.
- `query/`: handwritten SQL separated by domain module and consumed by sqlc.
- `generated/sqlc/`: generated pgx code; do not edit manually.
- `repository/<module>/`: PostgreSQL adapters for Auth, User, and Staff.
- `transaction/`: shared transaction lifecycle for Identity repositories.

Repositories bind feature queries with `queries.WithTx(tx)` inside
`WithinTransaction`; transaction ownership stays at the repository/use-case
boundary that needs atomicity.
