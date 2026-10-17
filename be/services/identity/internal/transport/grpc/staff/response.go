package staffgrpc

import (
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func staffResponse(value model.Staff) *identityV1.StaffResponse {
	// Map the domain staff member to the gRPC response.
	return &identityV1.StaffResponse{
		Id:        value.ID,
		Email:     value.Email,
		FullName:  value.FullName,
		IsAdmin:   value.IsAdmin,
		RoleId:    value.RoleID,
		RoleName:  value.RoleName,
		Status:    value.Status,
		CreatedAt: value.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
}

func staffListResponse(items []model.Staff) *identityV1.StaffListResponse {
	// Map all domain staff members to the list response.
	response := &identityV1.StaffListResponse{
		Staff: make([]*identityV1.StaffResponse, 0, len(items)),
	}
	for _, item := range items {
		response.Staff = append(response.Staff, staffResponse(item))
	}
	return response
}
