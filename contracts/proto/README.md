# Internal gRPC contracts

These versioned Protocol Buffer files define private synchronous calls between
Gateway processes and owning services. They are not public HTTP APIs.

The generated Go client/server for the current Identity contract lives under
`contracts/gen/identity/v1` and is generated from the versioned proto source.
The shared contract module owns transport types only; business implementation
remains in the owning service. Public
HTTP/JSON contracts remain under `contracts/openapi/` and are owned by the
Gateway exposing them.

`identity/v1/admin.proto` reserves the private Staff/Role/Permission and admin
auth contract. Its generated client/server is transport-only; authorization and
persistence remain owned by the Identity service.

Identity currently defines Auth, User CRUD, and Staff CRUD contracts under `identity/v1`; transport handlers remain owned by Identity.
