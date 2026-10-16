# Customer Booking Flow

## Actor

Customer, Payment Provider, Consumer, Cron Scheduler.

## Flow

1. The customer searches by location, date, and time.
2. The system returns available courts and server-calculated prices.
3. The customer creates a five-minute hold.
4. The database prevents overlapping active reservations.
5. The customer starts a mock payment.
6. The payment provider sends a callback.
7. The system processes the callback idempotently.
8. The booking changes to `CONFIRMED`.
9. An outbox event is published to Kafka.
10. The notification consumer informs the customer and venue.

## Failure Paths

- Slot conflict: return `BOOKING_SLOT_UNAVAILABLE`.
- Payment failure: mark `PAYMENT_FAILED`.
- Hold timeout: mark `EXPIRED` and release availability.
- Duplicate callback: return success without confirming twice.
