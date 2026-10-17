# Documentation Index

## Analysis

1. [Business Context](analysis/00-business-context.md)
2. [Owner Onboarding](analysis/01-owner-onboarding.md)
3. [Venue Setup](analysis/02-venue-setup.md)
4. [Customer Booking](analysis/03-customer-booking.md)
5. [Walk-In Booking](analysis/04-walk-in-booking.md)
6. [Cancellation, Refund, and Payout](analysis/05-cancellation-refund-payout.md)

## Design

1. [High-Level Design](design/01-high-level-design.md)
2. [Use Case Diagram](design/02-use-case-diagram.md)
3. [Core Flow Diagrams](design/03-core-flow-diagrams.md)
4. [Booking State Machine](design/04-booking-state-machine.md)
5. [Folder Structure](design/05-folder-structure.md)
6. [Microservice Migration Proposal](design/06-microservice-migration.md)

## Architecture Decisions

1. [Use gRPC for Internal Service Communication](adr/0001-grpc-for-internal-service-communication.md)
2. [Use Service-Scoped Storage Configuration](adr/0002-service-scoped-storage-configuration.md)
3. [Use an Independent Go Module per Service](adr/0003-independent-go-module-per-service.md)

## AI Context

- Start with [AI Context Index](ai-context/INDEX.md) before implementing or changing the project.
- Shared engineering rules are in [Project Conventions](09-project-conventions.md).

## Detailed References

The numbered documents in the parent `docs` directory remain the detailed source for requirements, data model, API contracts, testing, DevOps, and delivery planning.

- [Backend Engineering Practices](08-backend-engineering-practices.md) - correctness, concurrency, idempotency, failure handling, observability, security, testing, and review checklist.

## Operations

- [Release Runbook](../devops/RELEASE.md) - Jenkins, two-VM deployment, backup, rollback, and Kubernetes adoption path.
