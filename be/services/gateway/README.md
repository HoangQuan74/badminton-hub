# Gateway Service

The Gateway is the only public HTTP entry point. It owns route compatibility,
request/correlation metadata, forwarding, edge readiness, and rate-limit hooks.
It does not own business tables or domain state.

Gateway is an independent Go module (`go.mod`/`go.sum`) and imports only the
infrastructure-only `platform` module plus its own packages. It does not import
Identity code; runtime calls use the documented service contract.

Gateway is stateless in the current slice, but any edge cache, rate-limit, or
runtime Redis dependency must use `GATEWAY_REDIS_URL`. It must never reuse a
business service's database or Redis configuration.
