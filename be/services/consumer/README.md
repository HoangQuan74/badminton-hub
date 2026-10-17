# Consumer Process

Owns outbox publishing, Kafka consumption, notification commands, and selected
Kafka-to-Redis realtime fan-out. It does not own authoritative business state.

Realtime fan-out uses `CONSUMER_REDIS_URL`. Outbox rows remain owned by the
service that committed them; a consumer must use an explicit owner-scoped
connection or event contract rather than another service's schema.

## Local Kafka

Start the local Kafka-compatible broker and its topics from the repository root:

```bash
docker compose -f devops/docker/compose.yml up -d kafka kafka-init
```

The broker is available to host processes at `localhost:19093` and to Compose
peers at `kafka:9092`. The bootstrap process creates `booking.events`,
`payment.events`, `notification.commands`, and `payout.events`, each with three
partitions. Use `CONSUMER_KAFKA_BROKERS` and `CONSUMER_KAFKA_GROUP_ID` for the
Consumer process; consumers must remain idempotent because delivery is
at-least-once.
