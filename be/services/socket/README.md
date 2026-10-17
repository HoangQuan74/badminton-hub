# Socket Process

Owns authenticated WebSocket connections, Redis Pub/Sub subscription, and local
fan-out. Socket does not own durable business state.

Socket uses its own `SOCKET_REDIS_URL` for Pub/Sub and presence. It never reads
another service's schema or Redis namespace directly.

The Socket Gateway now connects to this process through
`SOCKET_GATEWAY_GRPC_TARGET`; the former HTTP upstream bridge has been removed.
The versioned realtime stream contract is implemented under
`contracts/proto/socket/v1`; the Gateway bridges WSS frames to the gRPC stream.
