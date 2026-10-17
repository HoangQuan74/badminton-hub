package staffgrpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/auth"
	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/security"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *StaffServer) Login(ctx context.Context, request *identityV1.AdminLoginRequest) (*identityV1.AdminAuthResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *identityV1.AdminLoginRequest) (*identityV1.AdminAuthResponse, error) {
		var id, email, fullName, passwordHash, status, roleID, roleName string
		var isAdmin bool
		err := s.authorizerPool().QueryRow(ctx, `SELECT u.id::text,u.email,u.full_name,u.password_hash,u.status,u.is_admin,COALESCE(u.role_id::text,''),COALESCE(r.name,'') FROM identity.users u LEFT JOIN identity.roles r ON r.id=u.role_id WHERE lower(u.email)=lower($1) AND (u.is_admin OR u.role_id IS NOT NULL)`, strings.TrimSpace(request.GetEmail())).Scan(&id, &email, &fullName, &passwordHash, &status, &isAdmin, &roleID, &roleName)
		if err != nil || !security.ComparePassword(passwordHash, request.GetPassword()) {
			return nil, auth.ErrInvalidCredentials
		}
		if status != "active" {
			return nil, auth.ErrUserInactive
		}
		access, accessExpiry, err := s.tokens.NewAccessToken(id)
		if err != nil {
			return nil, err
		}
		rawRefresh, refreshHash, refreshExpiry, err := s.tokens.NewRefreshToken()
		if err != nil {
			return nil, err
		}
		_, err = s.authorizerPool().Exec(ctx, `INSERT INTO identity.auth_sessions(id,user_id,refresh_token_hash,expires_at) VALUES($1,$2,$3,$4)`, uuid.NewString(), id, refreshHash, refreshExpiry)
		if err != nil {
			return nil, err
		}
		return &identityV1.AdminAuthResponse{Staff: &identityV1.StaffResponse{Id: id, Email: email, FullName: fullName, IsAdmin: isAdmin, RoleId: roleID, RoleName: roleName, Status: status}, Tokens: &identityV1.TokenPair{AccessToken: access, RefreshToken: rawRefresh, TokenType: "Bearer", AccessExpiresAt: accessExpiry.UTC().Format("2006-01-02T15:04:05Z07:00"), RefreshExpiresAt: refreshExpiry.UTC().Format("2006-01-02T15:04:05Z07:00")}}, nil
	})
}

func (s *StaffServer) Refresh(ctx context.Context, request *identityV1.AdminRefreshRequest) (*identityV1.AdminRefreshResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *identityV1.AdminRefreshRequest) (*identityV1.AdminRefreshResponse, error) {
		if request.GetRefreshToken() == "" {
			return nil, auth.ErrInvalidToken
		}
		var id string
		err := s.authorizerPool().QueryRow(ctx, `SELECT user_id::text FROM identity.auth_sessions s JOIN identity.users u ON u.id=s.user_id WHERE s.refresh_token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at>NOW() AND u.status='active' AND (u.is_admin OR u.role_id IS NOT NULL)`, auth.HashRefreshToken(request.GetRefreshToken())).Scan(&id)
		if err != nil {
			return nil, auth.ErrInvalidToken
		}
		access, accessExpiry, err := s.tokens.NewAccessToken(id)
		if err != nil {
			return nil, err
		}
		rawRefresh, refreshHash, refreshExpiry, err := s.tokens.NewRefreshToken()
		if err != nil {
			return nil, err
		}
		_, err = s.authorizerPool().Exec(ctx, `UPDATE identity.auth_sessions SET revoked_at=NOW() WHERE refresh_token_hash=$1`, auth.HashRefreshToken(request.GetRefreshToken()))
		if err != nil {
			return nil, err
		}
		_, err = s.authorizerPool().Exec(ctx, `INSERT INTO identity.auth_sessions(id,user_id,refresh_token_hash,expires_at) VALUES($1,$2,$3,$4)`, uuid.NewString(), id, refreshHash, refreshExpiry)
		if err != nil {
			return nil, err
		}
		return &identityV1.AdminRefreshResponse{Tokens: &identityV1.TokenPair{AccessToken: access, RefreshToken: rawRefresh, TokenType: "Bearer", AccessExpiresAt: accessExpiry.UTC().Format("2006-01-02T15:04:05Z07:00"), RefreshExpiresAt: refreshExpiry.UTC().Format("2006-01-02T15:04:05Z07:00")}}, nil
	})
}

func (s *StaffServer) Logout(ctx context.Context, request *identityV1.AdminLogoutRequest) (*identityV1.AdminLogoutResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *identityV1.AdminLogoutRequest) (*identityV1.AdminLogoutResponse, error) {
		if request.GetRefreshToken() == "" {
			return nil, auth.ErrInvalidToken
		}
		_, err := s.authorizerPool().Exec(ctx, `UPDATE identity.auth_sessions SET revoked_at=NOW() WHERE refresh_token_hash=$1 AND revoked_at IS NULL`, auth.HashRefreshToken(request.GetRefreshToken()))
		return &identityV1.AdminLogoutResponse{}, err
	})
}
func (s *StaffServer) GetCurrentStaff(ctx context.Context, request *identityV1.CurrentStaffRequest) (*identityV1.StaffResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *identityV1.CurrentStaffRequest) (*identityV1.StaffResponse, error) {
		item, err := s.authorizer.RequireAdmin(ctx, request.GetAccessToken())
		if err != nil {
			return nil, err
		}
		return staffResponse(item), nil
	})
}
func (s *StaffServer) ChangePassword(ctx context.Context, request *identityV1.ChangePasswordRequest) (*identityV1.Empty, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *identityV1.ChangePasswordRequest) (*identityV1.Empty, error) {
		item, err := s.authorizer.RequireAdmin(ctx, request.GetAccessToken())
		if err != nil {
			return nil, err
		}
		var oldHash string
		if err = s.authorizerPool().QueryRow(ctx, `SELECT password_hash FROM identity.users WHERE id=$1`, item.ID).Scan(&oldHash); err != nil {
			return nil, auth.ErrUserNotFound
		}
		if !security.ComparePassword(oldHash, request.GetCurrentPassword()) {
			return nil, auth.ErrInvalidCredentials
		}
		next, err := security.HashPassword(request.GetNewPassword())
		if err != nil {
			return nil, err
		}
		_, err = s.authorizerPool().Exec(ctx, `UPDATE identity.users SET password_hash=$2,updated_at=NOW() WHERE id=$1`, item.ID, next)
		return &identityV1.Empty{}, err
	})
}

// authorizerPool is supplied by the authorizer's repository boundary.
func (s *StaffServer) authorizerPool() *pgxpool.Pool { return s.pool }

func (s *StaffServer) ListRoles(ctx context.Context, request *identityV1.ListRolesRequest) (*identityV1.RoleListResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, grpcEndpoint.Authorize[*identityV1.ListRolesRequest](s.authorizeToken), func(ctx context.Context, request *identityV1.ListRolesRequest) (*identityV1.RoleListResponse, error) {
		rows, err := s.pool.Query(ctx, `SELECT id::text,name,description,created_at FROM identity.roles ORDER BY name`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		response := &identityV1.RoleListResponse{}
		for rows.Next() {
			var role identityV1.Role
			if err := rows.Scan(&role.Id, &role.Name, &role.Description, &role.CreatedAt); err != nil {
				return nil, err
			}
			response.Roles = append(response.Roles, &role)
		}
		return response, rows.Err()
	})
}
func (s *StaffServer) CreateRole(ctx context.Context, request *identityV1.CreateRoleRequest) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, grpcEndpoint.Authorize[*identityV1.CreateRoleRequest](s.authorizeToken), func(ctx context.Context, request *identityV1.CreateRoleRequest) (*identityV1.RoleResponse, error) {
		name := strings.TrimSpace(request.GetName())
		if name == "" {
			return nil, staffDomain.ErrInvalidInput
		}
		var role identityV1.Role
		err := s.pool.QueryRow(ctx, `INSERT INTO identity.roles(id,name,description) VALUES($1,$2,$3) RETURNING id::text,name,description,created_at`, uuid.NewString(), name, strings.TrimSpace(request.GetDescription())).Scan(&role.Id, &role.Name, &role.Description, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		return &identityV1.RoleResponse{Role: &role}, nil
	})
}
func (s *StaffServer) GetRole(ctx context.Context, request *identityV1.GetRoleRequest) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, grpcEndpoint.Authorize[*identityV1.GetRoleRequest](s.authorizeToken), func(ctx context.Context, request *identityV1.GetRoleRequest) (*identityV1.RoleResponse, error) {
		if _, err := uuid.Parse(request.GetRoleId()); err != nil {
			return nil, staffDomain.ErrInvalidInput
		}
		var role identityV1.Role
		err := s.pool.QueryRow(ctx, `SELECT id::text,name,description,created_at FROM identity.roles WHERE id=$1`, request.GetRoleId()).Scan(&role.Id, &role.Name, &role.Description, &role.CreatedAt)
		if err == pgx.ErrNoRows {
			return nil, staffDomain.ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		return &identityV1.RoleResponse{Role: &role}, nil
	})
}
func (s *StaffServer) UpdateRole(ctx context.Context, request *identityV1.UpdateRoleRequest) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, grpcEndpoint.Authorize[*identityV1.UpdateRoleRequest](s.authorizeToken), func(ctx context.Context, request *identityV1.UpdateRoleRequest) (*identityV1.RoleResponse, error) {
		if _, err := uuid.Parse(request.GetRoleId()); err != nil {
			return nil, staffDomain.ErrInvalidInput
		}
		var role identityV1.Role
		err := s.pool.QueryRow(ctx, `UPDATE identity.roles SET name=$2,description=$3,updated_at=NOW() WHERE id=$1 RETURNING id::text,name,description,created_at`, request.GetRoleId(), strings.TrimSpace(request.GetName()), strings.TrimSpace(request.GetDescription())).Scan(&role.Id, &role.Name, &role.Description, &role.CreatedAt)
		if err == pgx.ErrNoRows {
			return nil, staffDomain.ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		return &identityV1.RoleResponse{Role: &role}, nil
	})
}
func (s *StaffServer) DeleteRole(ctx context.Context, request *identityV1.DeleteRoleRequest) (*identityV1.Empty, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, grpcEndpoint.Authorize[*identityV1.DeleteRoleRequest](s.authorizeToken), func(ctx context.Context, request *identityV1.DeleteRoleRequest) (*identityV1.Empty, error) {
		if _, err := uuid.Parse(request.GetRoleId()); err != nil {
			return nil, staffDomain.ErrInvalidInput
		}
		result, err := s.pool.Exec(ctx, `DELETE FROM identity.roles WHERE id=$1`, request.GetRoleId())
		if err != nil {
			return nil, err
		}
		if result.RowsAffected() == 0 {
			return nil, staffDomain.ErrNotFound
		}
		return &identityV1.Empty{}, nil
	})
}
func (s *StaffServer) ReplaceRolePermissions(ctx context.Context, request *identityV1.ReplaceRolePermissionsRequest) (*identityV1.RoleResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, grpcEndpoint.Authorize[*identityV1.ReplaceRolePermissionsRequest](s.authorizeToken), func(ctx context.Context, request *identityV1.ReplaceRolePermissionsRequest) (*identityV1.RoleResponse, error) {
		if _, err := uuid.Parse(request.GetRoleId()); err != nil {
			return nil, staffDomain.ErrInvalidInput
		}
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		if _, err = tx.Exec(ctx, `DELETE FROM identity.role_permissions WHERE role_id=$1`, request.GetRoleId()); err != nil {
			return nil, err
		}
		for _, permissionID := range request.GetPermissionIds() {
			if _, err = uuid.Parse(permissionID); err != nil {
				return nil, staffDomain.ErrInvalidInput
			}
			if _, err = tx.Exec(ctx, `INSERT INTO identity.role_permissions(role_id,permission_id) VALUES($1,$2)`, request.GetRoleId(), permissionID); err != nil {
				return nil, err
			}
		}
		var role identityV1.Role
		if err = tx.QueryRow(ctx, `SELECT id::text,name,description,created_at FROM identity.roles WHERE id=$1`, request.GetRoleId()).Scan(&role.Id, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &identityV1.RoleResponse{Role: &role}, nil
	})
}
func (s *StaffServer) ListPermissions(ctx context.Context, request *identityV1.ListPermissionsRequest) (*identityV1.PermissionListResponse, error) {
	return grpcEndpoint.Run(ctx, request, errorMapper, grpcEndpoint.Authorize[*identityV1.ListPermissionsRequest](s.authorizeToken), func(ctx context.Context, request *identityV1.ListPermissionsRequest) (*identityV1.PermissionListResponse, error) {
		rows, err := s.pool.Query(ctx, `SELECT id::text,module,path,action,description FROM identity.permissions ORDER BY module,path,action`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		response := &identityV1.PermissionListResponse{}
		for rows.Next() {
			var permission identityV1.Permission
			if err := rows.Scan(&permission.Id, &permission.Module, &permission.Path, &permission.Action, &permission.Description); err != nil {
				return nil, err
			}
			response.Permissions = append(response.Permissions, &permission)
		}
		return response, rows.Err()
	})
}
