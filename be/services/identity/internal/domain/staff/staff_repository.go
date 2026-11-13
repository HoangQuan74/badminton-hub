package staff

import (
	"context"

	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

type CreateStaffParams struct {
	Email, PasswordHash, FullName, RoleID string
	IsAdmin                               bool
}
type StaffRepository interface {
	ListStaff(ctx context.Context, query model.ListQuery) ([]model.Staff, error)
	GetStaff(ctx context.Context, staffID string) (model.Staff, error)
	CreateStaff(ctx context.Context, params CreateStaffParams) (model.Staff, error)
	UpdateStaff(ctx context.Context, input UpdateStaffInput) (model.Staff, error)
	UpdateStaffAccess(ctx context.Context, input UpdateStaffAccessInput) (model.Staff, error)
	DeleteStaff(ctx context.Context, staffID string) error
}

type StaffMembershipRepository interface {
	WithinMembershipTransaction(ctx context.Context, execute func(StaffMembershipTransaction) error) error
	ListOrganizationStaff(ctx context.Context, actorID, organizationID string) ([]model.OrganizationStaffDetails, error)
	SetOrganizationStaff(ctx context.Context, actorID string, input SetOrganizationStaffInput) (model.OrganizationStaff, error)
	SetVenueStaff(ctx context.Context, actorID string, input SetVenueStaffInput) (model.VenueStaff, error)
}

type StaffAccountRepository interface {
	FindStaffCredentialsByEmail(ctx context.Context, email string) (model.StaffCredentials, error)
	CreateStaffSession(ctx context.Context, session model.Session) error
	RotateStaffSession(ctx context.Context, tokenHash []byte, next model.Session) (string, error)
	RevokeStaffSession(ctx context.Context, tokenHash []byte) error
	GetStaffPasswordHash(ctx context.Context, staffID string) (string, error)
	UpdateStaffPassword(ctx context.Context, staffID, passwordHash string) error
}

type StaffMembershipTransaction interface {
	GetManagedOrganizationStatus(ctx context.Context, actorID, organizationID string) (string, error)
	LockCandidate(ctx context.Context, email string) (model.User, error)
	AssignRole(ctx context.Context, userID, role string) error
	SetOrganizationStaff(ctx context.Context, actorID string, input SetOrganizationStaffInput) (model.OrganizationStaff, error)
	DeactivateOrganizationStaff(ctx context.Context, actorID string, input RemoveOrganizationStaffInput) (model.OrganizationStaff, error)
	DeactivateMemberVenues(ctx context.Context, organizationID, userID string) error
}
