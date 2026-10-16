# Business Context

## Product

BadmintonHub is a B2B2C marketplace and venue-management platform.

- Customers search and book badminton courts.
- Owners manage venues, courts, schedules, prices, staff, and revenue.
- Staff manage daily operations and walk-in bookings.
- Platform admins approve venues and oversee payments and payouts.

## Revenue Model

- Online booking: the platform charges a commission.
- Walk-in booking: no platform commission in the MVP.
- Future option: monthly SaaS plans for venue owners.

## Core Rule

Online and walk-in bookings must share one calendar. PostgreSQL is the source of truth for availability and conflict prevention.
