# Consumer Process

Owns outbox publishing, Kafka consumption, notification commands, and selected
Kafka-to-Redis realtime fan-out. It does not own authoritative business state.

Realtime fan-out uses `CONSUMER_REDIS_URL`. Outbox rows remain owned by the
service that committed them; a consumer must use an explicit owner-scoped
connection or event contract rather than a shared business database.
