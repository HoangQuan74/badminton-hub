# Product Requirements

## Problem

Badminton venues often manage online messages, phone reservations, and walk-in customers separately. This creates stale availability, double booking, manual reconciliation, and limited visibility into court utilization and revenue.

BadmintonHub provides one source of truth for public discovery, online booking, walk-in booking, court schedules, payments, and venue operations.

## Users

### Customer

- Registers and signs in.
- Searches venues by location, date, time, price, and amenities.
- Views court availability and pricing.
- Holds, pays for, cancels, and reviews bookings.
- Receives booking and payment notifications.

### Venue Owner

- Creates an organization and submits it for approval.
- Creates one or more venues and courts.
- Configures operating hours, pricing, cancellation policy, and maintenance blocks.
- Invites staff and assigns venue access.
- Reviews bookings, revenue, refunds, and payout statements.

### Venue Staff

- Views the daily schedule for assigned venues.
- Creates walk-in or phone bookings.
- Checks customers in.
- Cancels or adjusts bookings within granted permissions.

### Platform Admin

- Reviews organizations and venues.
- Suspends accounts or venues.
- Reviews bookings, payments, refunds, disputes, and audit logs.
- Configures commission and approves venue payouts.

## MVP Functional Requirements

### Identity and access

- Email and password registration.
- Access and refresh tokens.
- Password hashing and refresh-token revocation.
- Roles: customer, owner, staff, admin.
- Organization and venue membership checks on every protected operation.

### Venue discovery

- Browse and search published venues.
- View venue address, location, images, amenities, opening hours, cancellation policy, and courts.
- Query available courts by date, start time, end time, and duration.

### Venue management

- Create and update organizations, venues, and courts.
- Configure weekly operating hours and exceptions.
- Configure weekday, weekend, and peak-hour prices.
- Block a court for maintenance or private use.
- Publish a venue only after platform approval.

### Booking

- Create a temporary court hold.
- Confirm a booking after successful payment.
- View booking details and history.
- Cancel according to the venue policy.
- Create walk-in bookings from the staff portal.
- Prevent overlap between online, walk-in, and maintenance blocks.

### Payment

- Create a mock payment intent.
- Simulate success, failure, and timeout.
- Process signed payment callbacks.
- Handle duplicated callbacks idempotently.
- Calculate platform commission and venue receivable.
- Record a mock refund when cancellation policy allows it.

### Operations

- Daily venue schedule.
- Booking check-in.
- Booking and payment audit history.
- Failed asynchronous-event visibility and manual retry.

## Core Business Rules

### Time and availability

- All persisted timestamps use UTC.
- Venue configuration stores an IANA timezone, initially `Asia/Ho_Chi_Minh`.
- A booking must be inside venue operating hours.
- Booking duration uses a venue-defined interval, initially 30 minutes.
- Past slots cannot be held or booked.
- A court block and an active booking cannot overlap.

Two ranges overlap when:

```text
new_start < existing_end AND new_end > existing_start
```

### Hold

- A hold lasts five minutes by default.
- The hold belongs to one customer and one idempotency key.
- An expired hold cannot be paid.
- Expiration releases availability through the Cron process and database truth, not only Redis TTL.

### Booking state

```text
PENDING_PAYMENT -> CONFIRMED -> CHECKED_IN -> COMPLETED
PENDING_PAYMENT -> PAYMENT_FAILED
PENDING_PAYMENT -> EXPIRED
CONFIRMED       -> CANCELLED
```

- The server owns state transitions.
- Every transition records actor, timestamp, previous state, next state, and reason.

### Pricing

- The server calculates price; the frontend never supplies the final amount.
- A booking stores a pricing snapshot.
- Later price changes do not affect existing bookings.
- MVP supports weekday, weekend, and peak-hour rules.

### Cancellation

Initial policy:

- More than 24 hours before start: 100% refund.
- Between 6 and 24 hours: 50% refund.
- Less than 6 hours: no refund.
- Venue cancellation: 100% refund.

The exact applied rule and financial breakdown are stored on the booking.

## Advanced Backlog

- Real-time availability updates.
- QR check-in.
- Waitlist.
- Recurring bookings.
- Customer ratings.
- Membership and loyalty.
- Player matching.
- Venue-branded subdomains.
- Tournament management.
- Dynamic pricing.

## Non-Goals for Week One

- Real payment-provider settlement.
- Customer wallet.
- Native mobile application.
- Custom domains.
- Separate deployment or database per tenant.
- Kubernetes.
- Complex accounting or tax compliance.
