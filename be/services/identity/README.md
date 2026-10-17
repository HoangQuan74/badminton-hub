# Identity Service

Identity owns the `identity` PostgreSQL schema, user and staff accounts,
credentials, refresh sessions, JWT issuance, admin roles/permissions, and
private gRPC methods. Its PostgreSQL adapters, SQLC output, and service-local
helpers stay inside this module boundary.

## Layout

```text
identity/
  cmd/identity/                         # process composition and lifecycle
  config/                                # IDENTITY_* configuration
  internal/
    domain/
      auth/                              # user credentials, sessions, tokens
      user/                              # User CRUD rules and use-case inputs
      staff/                             # Staff CRUD rules and access inputs
    model/                                # entities shared by domain, storage, and gRPC
      user.go                             # reusable User entity
      staff.go                            # reusable Staff entity
      list_query.go                       # shared pagination, search, and date filters
    security/                            # password hashing
    validation/                          # email and input normalization
    storage/postgres/
      schema/
        migrations/                      # identity schema migrations
        seed/                            # local development seed
      query/                             # SQL source split by module
        auth.sql
        user.sql
        staff.sql
      generated/sqlc/                    # generated PostgreSQL code
      repository/
        auth/                            # Auth/session persistence adapter
        user/                            # User CRUD persistence adapter
        staff/                           # Staff CRUD/access persistence adapter
      transaction/                       # shared transaction lifecycle
    transport/grpc/auth/                 # private Auth gRPC server
    transport/grpc/user/                 # private User CRUD gRPC server
    transport/grpc/staff/                # private Staff CRUD gRPC server
  test/                                  # service-owned test documentation/fixtures
```

The public entry point is a Gateway. Gateway code does not import this module;
it calls the versioned contracts under `contracts/proto` over gRPC. The HTTP
transport remains only while the Admin and Socket gateway migrations are being
completed.

Identity uses `IDENTITY_DATABASE_URL`, `IDENTITY_REDIS_URL`,
`IDENTITY_GRPC_ADDR`, and namespaced timeout/token settings. It never reads
another service's schema or migration directory.
