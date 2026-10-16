# Business Model

## Model

BadmintonHub is a B2B2C marketplace combined with a lightweight SaaS portal for venue operators.

- B2C: customers discover venues, reserve courts, pay, cancel, and check in.
- B2B: owners manage courts, schedules, staff, walk-in bookings, and revenue.
- Platform: BadmintonHub provides discovery, booking infrastructure, payment records, operational controls, and reporting.

## Revenue

The MVP uses a commission model for online bookings.

```text
court amount        = 200,000 VND
discount            =   0,000 VND
customer paid       = 200,000 VND
platform fee (5%)   =  10,000 VND
venue receivable    = 190,000 VND
```

Walk-in bookings are recorded in the same calendar but do not incur platform commission in the MVP.

Future revenue options:

- Monthly owner subscriptions.
- Featured venue placement.
- Payment processing fees.
- Membership and loyalty packages.
- Equipment rental and coaching add-ons.
- Tournament-management fees.

## Marketplace Lifecycle

```text
Owner registers organization
-> Admin approves organization
-> Owner creates venue and courts
-> Admin approves venue
-> Venue becomes publicly searchable
-> Customer holds and pays for a court
-> Staff checks customer in
-> Booking completes
-> Platform calculates commission
-> Admin approves mock venue payout
```

## Organization Lifecycle

```text
PENDING_VERIFICATION -> ACTIVE -> SUSPENDED
PENDING_VERIFICATION -> REJECTED
```

## Venue Lifecycle

```text
DRAFT -> PENDING_APPROVAL -> PUBLISHED -> SUSPENDED
```

## Booking Sources

### Online

- Created by a customer.
- Requires a hold and payment.
- Platform commission applies.

### Walk-in or phone

- Created by venue staff.
- Customer account is optional.
- Must use the same availability and conflict protection as online booking.
- No platform commission in the MVP.

## Multi-Tenant Model

- One owner can manage multiple organizations only if explicitly added as a member.
- One organization can own multiple venues.
- One venue contains multiple courts.
- Staff access is restricted to assigned organizations or venues.
- Customer-facing venue pages use public slugs such as `/venues/smash-center`.

The MVP uses one shared product surface, while each deployed service owns its
database. Tenant isolation is enforced by organization-scoped authorization and
queries inside the owning service. Custom subdomains and dedicated deployments
are future options, not MVP requirements.
