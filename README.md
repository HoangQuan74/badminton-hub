# BadmintonHub

BadmintonHub is a multi-tenant badminton court marketplace and venue-management platform. Customers discover venues and book courts, while venue owners manage courts, schedules, pricing, staff, bookings, and operational reports.

This repository contains the product documentation and the first backend and customer-web implementation slices.

## Product Model

- Public marketplace for customers.
- Multi-tenant owner and staff portal.
- Platform administration portal.
- Commission on online bookings.
- Walk-in bookings managed by venue staff.
- Tenant-specific public pages such as `/venues/smash-center`.
- One shared PostgreSQL deployment with one owner schema per stateful service
  (`identity`, `catalog`, `booking`, `payment`). Stateless processes do not own
  business tables. Organization isolation remains enforced inside the owning
  service schema.

## Planned Stack

- Frontend: React, Vite, TypeScript, React Router, TanStack Query, Zustand, React Hook Form, Zod.
- Backend: Go, Gin, PostgreSQL, Redis, Kafka, WebSocket.
- Testing: Go table-driven tests, Testify, GoConvey, integration and end-to-end tests.
- DevOps: Docker Compose, Jenkins, Prometheus, Grafana, AWS, Nginx or Caddy.

## Documentation

Start with [Documentation Index](docs/README.md). AI-assisted work starts with [AI Context Index](docs/ai-context/INDEX.md).

1. [Product Requirements](docs/01-product-requirements.md)
2. [Business Model](docs/02-business-model.md)
3. [System Design](docs/03-system-design.md)
4. [Data Model](docs/04-data-model.md)
5. [API and Event Contracts](docs/05-api-and-events.md)
6. [Testing and DevOps](docs/06-testing-and-devops.md)
7. [Seven-Day Delivery Plan](docs/07-delivery-plan.md)
8. [Backend Engineering Practices](docs/08-backend-engineering-practices.md)

## MVP Success Scenarios

The MVP is complete only when it demonstrates these scenarios:

1. Two customers try to reserve the same court and time; only one succeeds.
2. A duplicated payment callback confirms the booking only once.
3. An unpaid hold expires and releases the court automatically.
4. Online and walk-in bookings use the same availability source and cannot overlap.
5. Jenkins validates, builds, and deploys the application to staging.
6. Prometheus and Grafana show API, Socket, Consumer, Cron, booking, and payment health.

## Scope Rule

Correct booking, payment, tenant isolation, tests, and operability take priority over feature count. Tournament management, player matching, subscriptions, complex loyalty, and Kubernetes are not part of the first-week MVP.
