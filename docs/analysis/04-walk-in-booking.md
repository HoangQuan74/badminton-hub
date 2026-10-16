# Walk-In Booking Flow

## Actor

Venue Staff, Customer.

## Flow

1. Staff opens the venue calendar.
2. Staff selects a court and time range.
3. Staff enters customer name and phone number.
4. The system checks the same availability rules as online booking.
5. The booking is created with source `WALK_IN`.
6. The court becomes unavailable to online customers.
7. Staff checks the customer in when they arrive.

## Rules

- A customer account is optional.
- Staff can access only assigned venues.
- Walk-in bookings use database conflict protection.
- Platform commission is zero for the MVP.
- Creation, cancellation, and check-in are audited.
