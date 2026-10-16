# Socket Process

Owns authenticated WebSocket connections, Redis Pub/Sub subscription, and local
fan-out. Socket does not own durable business state.

Socket uses its own `SOCKET_REDIS_URL` for Pub/Sub and presence. It never reads
another service's database or Redis namespace directly.
