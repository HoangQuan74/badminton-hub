# Identity Service

Identity owns the `identity` PostgreSQL schema, user and staff accounts,
credentials, refresh sessions, JWT issuance, admin roles/permissions, tenant
organizations, venues, staff memberships and venue assignments. Its PostgreSQL adapters, SQLC output, and service-local
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
      role/                              # role and permission management
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
        role.sql
      generated/sqlc/                    # generated PostgreSQL code
      repository/
        auth/                            # Auth/session persistence adapter
        user/                            # User CRUD persistence adapter
        staff/                           # Staff CRUD/access persistence adapter
        role/                            # Role and permission persistence adapter
      transaction/                       # shared transaction lifecycle
    transport/grpc/auth/                 # admin, tenant, and user auth servers
    transport/grpc/user/                 # private User CRUD gRPC server
    transport/grpc/staff/                # staff CRUD handlers
    transport/grpc/role/                 # private role and permission handlers
  test/                                  # service-owned test documentation/fixtures
```

The public entry point is a Gateway. Gateway code does not import this module;
it calls the versioned contracts under `contracts/proto` over gRPC. The HTTP
transport remains only while the Admin and Socket gateway migrations are being
completed.

Identity uses `IDENTITY_DATABASE_URL`, `IDENTITY_REDIS_URL`,
`IDENTITY_GRPC_ADDR`, and namespaced timeout/token settings. It never reads
another service's schema or migration directory.

Authentication, refresh sessions and password changes for both user and admin
accounts belong to `auth`. The `user` and `staff` modules manage their CRUD
operations. Admin Gateway uses the separate `IdentityAdminAuthService`,
`IdentityStaffService`, and `IdentityRoleService` contracts. Identity also
registers the existing `IdentityAdminService` as a compatibility endpoint for
older gateways. Deploy Identity before updating the Admin Gateway; the new
gateway requires the separate service endpoints.

## Tenant and venue access

Active `tenant_staff` members can read venues throughout their tenant. Active
`venue_staff` members can read only venues with an active assignment. Tenant
owners manage organizations, venues, staff membership and venue assignments.
The staff assignment endpoints operate on existing active users with the matching
global role; they do not create users or send invitations. Identity migration
`000003` provisions the `venue_staff` role independently of local seeds.
See `devops/README.md` for migration and rollback requirements.

## Authentication groups

Authentication is organized into three flows under `transport/grpc/auth`:

- `admin_handler.go` / `AdminServer`: platform administrators.
- `tenant_handler.go` / `TenantServer`: tenant owners, tenant staff, and venue staff.
- `user_handler.go` / `UserServer`: customer registration and authentication.

`admin_server.go`, `tenant_server.go`, and `user_server.go` wire the respective
dependencies. Actor prefixes keep each flow together when sorting by filename. Domain implementations live in
`domain/auth/admin_service.go`, `tenant_service.go`, and `user_service.go`.
Token generation and credential/session persistence are shared. Tenant auth
always enforces the tenant portal. The existing `IdentityAuthService` continues
to accept the legacy portal field for older gateways; current tenant routes use
`IdentityTenantAuthService`. Deploy Identity before updating User Gateway.
