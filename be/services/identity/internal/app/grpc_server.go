package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	identityConfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	authDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	permissionDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/permission"
	roleDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/role"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	tenantDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	userDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/user"
	venueDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/venue"
	authPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/auth"
	permissionPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/permission"
	rolePostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/role"
	staffPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/staff"
	tenantPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/tenant"
	userPostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/user"
	venuePostgres "github.com/kainguyen/badminton-hub/be/services/identity/internal/repository/postgres/venue"
	identityGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc"
	authGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/auth"
	permissionGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/permission"
	roleGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/role"
	staffGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/staff"
	tenantGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/tenant"
	userGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/user"
	venueGRPC "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/venue"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc"
)

func RegisterGRPCServices(cfg identityConfig.Config, pool *pgxpool.Pool) func(*grpc.Server) {
	// Compose service-owned infrastructure through domain repository interfaces.
	tokens := authDomain.NewTokenManager(
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)
	authRepository := authPostgres.NewAuthRepository(pool)
	staffRepository := staffPostgres.NewRepository(pool)
	authService := authDomain.NewService(
		authRepository,
		tokens,
		userDomain.CanAccessPortal,
	)
	staffService := staffDomain.NewService(staffRepository)
	membershipService := staffDomain.NewMembershipService(staffRepository)
	authorizer := identityGRPC.NewAuthorizer(tokens, staffService)

	// Inject domain services into protocol adapters.
	userAuthServer := authGRPC.NewUserServer(authService, userDomain.NewRegistrationService(userPostgres.NewRepository(pool), tokens))
	tenantAuthServer := authGRPC.NewTenantServer(authDomain.NewService(
		authRepository,
		tokens,
		tenantDomain.CanAccessPortal,
	))
	adminAuthServer := authGRPC.NewAdminServer(staffDomain.NewAccountService(staffRepository, tokens), authorizer)
	staffServer := staffGRPC.NewStaffServer(
		staffService,
		membershipService,
		authService,
		authorizer,
	)
	roleServer := roleGRPC.NewRoleServer(roleDomain.NewService(rolePostgres.NewRepository(pool)), authorizer)
	permissionServer := permissionGRPC.NewPermissionServer(permissionDomain.NewService(permissionPostgres.NewRepository(pool)), authorizer)
	userServer := userGRPC.NewUserServer(userDomain.NewService(userPostgres.NewRepository(pool)), authorizer)
	tenantServer := tenantGRPC.NewTenantServer(
		tenantDomain.NewService(tenantPostgres.NewRepository(pool)),
		authService,
		tokens,
		authorizer,
	)
	venueServer := venueGRPC.NewVenueServer(venueDomain.NewService(venuePostgres.NewRepository(pool)), authService)
	return func(server *grpc.Server) {
		identityV1.RegisterIdentityAuthServiceServer(server, userAuthServer)
		identityV1.RegisterIdentityTenantAuthServiceServer(server, tenantAuthServer)
		identityV1.RegisterIdentityAdminAuthServiceServer(server, adminAuthServer)
		identityV1.RegisterIdentityStaffServiceServer(server, staffServer)
		identityV1.RegisterIdentityRoleServiceServer(server, roleServer)
		identityV1.RegisterIdentityPermissionServiceServer(server, permissionServer)
		identityV1.RegisterIdentityUserServiceServer(server, userServer)
		identityV1.RegisterIdentityTenantServiceServer(server, tenantServer)
		identityV1.RegisterIdentityVenueServiceServer(server, venueServer)
	}
}
