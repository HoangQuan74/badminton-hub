# Venue Setup Flow

## Actor

Venue Owner, Platform Admin.

## Flow

1. The owner creates a venue draft.
2. The owner adds courts and amenities.
3. The owner configures operating hours.
4. The owner creates pricing rules.
5. The owner defines slot duration and cancellation policy.
6. The owner submits the venue for approval.
7. The admin approves the venue.
8. The venue becomes searchable in the public marketplace.

## Rules

- A venue belongs to one organization.
- A court belongs to one venue.
- Pricing periods cannot overlap at the same priority.
- Court blocks remove time from public availability.
- Only `PUBLISHED` venues accept online bookings.
