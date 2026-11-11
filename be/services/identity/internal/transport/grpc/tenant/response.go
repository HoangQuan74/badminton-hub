package tenantgrpc

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func organizationResponse(item model.Organization) *identityV1.TenantOrganization {
	return &identityV1.TenantOrganization{
		Id:        item.ID,
		Name:      item.Name,
		Slug:      item.Slug,
		Status:    item.Status,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func venueResponse(item model.Venue) *identityV1.TenantVenue {
	return &identityV1.TenantVenue{
		Id:             item.ID,
		OrganizationId: item.OrganizationID,
		Name:           item.Name,
		Slug:           item.Slug,
		Address:        item.Address,
		Timezone:       item.Timezone,
		Status:         item.Status,
	}
}

func mapVenues(rows []model.Venue) []*identityV1.TenantVenue {
	items := make([]*identityV1.TenantVenue, 0, len(rows))
	for _, row := range rows {
		items = append(items, venueResponse(row))
	}
	return items
}

func organizationStaffResponse(item model.OrganizationStaff) *identityV1.TenantStaffResponse {
	return &identityV1.TenantStaffResponse{
		OrganizationId: item.OrganizationID,
		UserId:         item.UserID,
		Role:           item.Role,
		Status:         item.Status,
	}
}

func venueStaffResponse(item model.VenueStaff) *identityV1.TenantVenueStaffResponse {
	return &identityV1.TenantVenueStaffResponse{
		OrganizationId: item.OrganizationID,
		VenueId:        item.VenueID,
		UserId:         item.UserID,
		Status:         item.Status,
	}
}
