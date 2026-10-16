# Booking service

This is the Booking ownership boundary. When implemented, it will own its
handlers, domain rules, adapters, `BOOKING_DATABASE_URL`,
`BOOKING_REDIS_URL`, and the migrations under `migrations/`. It must
not query another service's database; use gRPC for synchronous calls and Kafka
events for asynchronous work.

The service is not deployed yet.
