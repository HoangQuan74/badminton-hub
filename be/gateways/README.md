# Gateway processes

Gateway processes are separate deployable edges. They share only infrastructure primitives from `be/platform`; they do not share
business handlers or storage.

| Process | Folder | Default port | Internal transport |
| --- | --- | ---: | --- |
| User Gateway | `gateways/user` | `8081` | Identity gRPC |
| Admin Gateway | `gateways/admin` | `8082` | Identity gRPC |
| Socket Gateway | `gateways/socket` | `8083` | Socket gRPC stream |

The User Gateway accepts the existing `/api/v1/auth/*` and `/api/v1/me` routes.
The Admin Gateway accepts `/api/v1/admin/*`. The Socket Gateway accepts `/ws`
and `/ws/*path`, bridging WebSocket frames to the Socket gRPC stream.

Each gateway also exposes its embedded contract and Swagger UI:

```text
/openapi.yaml
/docs/
```

The User Gateway documents public user HTTP operations, the Admin Gateway
documents its forwarded admin surface, and the Socket Gateway documents the
WebSocket upgrade endpoint.

User and Admin Gateways translate their public HTTP contracts to Identity gRPC.
Admin Staff/Auth/Role/Permission routes are wired to the IdentityAdminService
contract; Identity currently returns `FEATURE_NOT_IMPLEMENTED` for the modules
whose domain handlers are not implemented yet. Socket bridges WSS frames to the versioned Socket gRPC stream. No HTTP bridge remains.

The local Compose file keeps them opt-in so the customer stack remains small:

```bash
docker compose -f devops/docker/compose.yml --profile admin up --build admin-gateway
docker compose -f devops/docker/compose.yml --profile socket up --build socket-gateway
```

Run each process directly from the repository root:

```bash
(cd be/gateways/user && go run ./cmd/user-gateway)
(cd be/gateways/admin && go run ./cmd/admin-gateway)
(cd be/gateways/socket && go run ./cmd/socket-gateway)
```
