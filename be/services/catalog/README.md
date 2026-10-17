# Catalog service

This is the Catalog ownership boundary. When implemented, it will own the
`catalog` PostgreSQL schema, handlers, domain rules, adapters,
`CATALOG_DATABASE_URL`, `CATALOG_REDIS_URL`, and the migrations under
`migrations/`. It must not query another service's schema; use gRPC for synchronous calls and Kafka
events for asynchronous work.

The service is not deployed yet.
