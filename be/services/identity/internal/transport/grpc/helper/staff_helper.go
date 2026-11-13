package helper

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

// MapStaffResponse maps shared access fields. Timestamp inclusion belongs to the caller.
func MapStaffResponse(staff model.Staff) *identityV1.StaffResponse {
	return &identityV1.StaffResponse{
		Id:       staff.ID,
		Email:    staff.Email,
		FullName: staff.FullName,
		IsAdmin:  staff.IsAdmin,
		RoleId:   staff.RoleID,
		RoleName: staff.RoleName,
		Status:   staff.Status,
	}
}

func MapOrganizationStaffResponse(staff model.OrganizationStaff) *identityV1.TenantStaffResponse {
	return &identityV1.TenantStaffResponse{
		OrganizationId: staff.OrganizationID,
		UserId:         staff.UserID,
		Role:           staff.Role,
		Status:         staff.Status,
	}
}
