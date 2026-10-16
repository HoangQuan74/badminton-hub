# Booking State Machine

```mermaid
stateDiagram-v2
    [*] --> PENDING_PAYMENT: Hold created
    PENDING_PAYMENT --> CONFIRMED: Payment succeeds
    PENDING_PAYMENT --> PAYMENT_FAILED: Payment fails
    PENDING_PAYMENT --> EXPIRED: Hold timeout
    CONFIRMED --> CHECKED_IN: Staff checks in
    CHECKED_IN --> COMPLETED: Session ends
    CONFIRMED --> CANCELLED: Valid cancellation
    PAYMENT_FAILED --> [*]
    EXPIRED --> [*]
    CANCELLED --> [*]
    COMPLETED --> [*]
```

## Rules

- The server owns transitions.
- Every transition is validated and audited.
- Payment callbacks are idempotent.
- An expired hold cannot become confirmed automatically.
- A completed or cancelled booking is terminal.
