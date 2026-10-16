# Notification Service

Future boundary for notification policy and delivery commands. Durable event
consumption remains owned by `services/consumer`.

Notification has no authoritative business database. Any delivery deduplication
or short-lived state uses `NOTIFICATION_REDIS_URL`; it must not reuse another
service's storage configuration.
