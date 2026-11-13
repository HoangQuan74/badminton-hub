package venuegrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

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
