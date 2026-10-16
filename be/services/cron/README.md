# Cron Process

Owns scheduled hold expiration and payout jobs. Jobs use distributed locking
and idempotent state transitions; Cron does not keep a second business store.

Cron has no independent business database in the current design. Scheduled
commands execute through the owning service contract, and any temporary
coordination state uses `CRON_REDIS_URL`.
