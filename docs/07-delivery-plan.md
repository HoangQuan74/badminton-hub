# Seven-Day Delivery Plan

## Delivery Principle

Complete one correct end-to-end booking flow before expanding features. Each day ends with working, reviewable behavior and tests for the highest-risk rules.

## Day 1 - Foundation and Identity

- Create monorepo structure for web, API, Socket, Consumer, Cron, contracts, and infrastructure.
- Add Docker Compose for PostgreSQL, Redis, Kafka, Gateway, Identity, Socket, Consumer, and Cron.
- Add Identity database migrations and seed data under the Identity service
  boundary; future services add their own migration histories.
- Implement registration, login, refresh, and RBAC.
- Create organization, venue, court, and membership models.
- Publish initial OpenAPI contract.

Exit criteria: a user can register, create an organization, and access an owner-protected endpoint.

## Day 2 - Venue, Schedule, and Pricing

- Venue and court management.
- Operating hours.
- Maintenance blocks.
- Pricing rules and server-side price calculation.
- Public venue list and detail page.
- Availability query.

Exit criteria: an approved venue exposes valid available slots and prices.

## Day 3 - Booking Correctness

- Temporary hold.
- Booking state machine.
- PostgreSQL overlap protection.
- Idempotency keys.
- Concurrent booking integration test.
- Walk-in booking using the same conflict rules.

Exit criteria: two concurrent requests for the same court and time produce exactly one active reservation.

## Day 4 - Payment and Reliable Events

- Mock payment intent and callback.
- Callback signature and deduplication.
- Booking confirmation transaction.
- Outbox pattern.
- Kafka publishing and idempotent consumers.
- Kafka-to-Redis realtime dispatcher.
- Hold-expiration Cron job with distributed locking.

Exit criteria: duplicate payment callback confirms once, and unpaid holds release automatically.

## Day 5 - Full-Stack User Flows

- Customer search, availability, checkout, and booking history.
- Owner court and schedule management.
- Staff daily calendar and walk-in booking.
- Admin approval screens.
- Standalone Socket Service with JWT handshake, Redis Pub/Sub subscription, rooms, and local broadcast.

Exit criteria: customer and staff flows work through the browser without manual database changes.

## Day 6 - Quality and Operability

- Complete unit and integration tests.
- Prometheus metrics.
- Grafana dashboards.
- Structured logging and request IDs.
- Health, readiness, graceful shutdown, retry, and failed-event visibility.
- Postman collection.

Exit criteria: critical flows are tested and observable under success and failure.

## Day 7 - CI/CD and Portfolio Delivery

- Jenkins pipeline.
- Multi-stage Docker images.
- Container registry.
- AWS staging deployment.
- HTTPS and smoke test.
- Architecture diagram.
- Seeded demo accounts and data.
- Three-to-five-minute demo video.
- Final README with setup, tradeoffs, and known limitations.

Exit criteria: a reviewer can open the staging URL, run the repository locally, inspect a green pipeline, and reproduce the core scenarios.

## Scope Cut Order

If time is short, remove features in this order:

1. Owner analytics.
2. QR check-in.
3. Email delivery.
4. Live WebSocket availability.
5. Payout UI.

Never cut:

- Tenant authorization.
- Double-booking protection.
- Idempotent payment callback.
- Hold expiration.
- Critical integration tests.
- CI pipeline.
- Basic observability.
