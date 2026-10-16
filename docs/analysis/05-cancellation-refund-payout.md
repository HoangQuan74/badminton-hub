# Cancellation, Refund, and Payout Flow

## Actors

Customer, Venue Staff, Venue Owner, Platform Admin.

## Cancellation

1. An authorized actor requests cancellation.
2. The system checks booking status and start time.
3. The system selects the stored cancellation policy.
4. It calculates refund, cancellation fee, and financial reversal.
5. The booking becomes `CANCELLED`.
6. A refund record and outbox event are created atomically.

## Initial Refund Policy

- More than 24 hours: 100%.
- Between 6 and 24 hours: 50%.
- Less than 6 hours: 0%.
- Venue cancellation: 100%.

## Payout

1. Completed bookings are grouped by venue and period.
2. Refunds and platform commission are deducted.
3. The admin approves a mock payout.
4. The owner sees the payout statement.
5. The admin marks the payout as `PAID`.

The MVP records money movement but does not transfer real funds.
