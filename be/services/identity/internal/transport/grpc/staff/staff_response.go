package staffgrpc

import (
	"time"

	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func staffResponse(value model.Staff) *identityV1.StaffResponse {
	response := grpcHelper.MapStaffResponse(value)
	response.CreatedAt = value.CreatedAt.UTC().Format(time.RFC3339Nano)
	return response
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

func venueStaffResponse(item model.VenueStaff) *identityV1.TenantVenueStaffResponse {
	return &identityV1.TenantVenueStaffResponse{
		OrganizationId: item.OrganizationID,
		VenueId:        item.VenueID,
		UserId:         item.UserID,
		Status:         item.Status,
	}
}

func organizationStaffListResponse(items []model.OrganizationStaffDetails) *identityV1.OrganizationStaffListResponse {
	response := &identityV1.OrganizationStaffListResponse{Items: make([]*identityV1.OrganizationStaffMember, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, &identityV1.OrganizationStaffMember{
			OrganizationId: item.OrganizationID,
			UserId:         item.UserID,
			Role:           item.Role,
			Status:         item.Status,
			Email:          item.Email,
			FullName:       item.FullName,
			VenueIds:       item.VenueIDs,
		})
	}

	return response
}
