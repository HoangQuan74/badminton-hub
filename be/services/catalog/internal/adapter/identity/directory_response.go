package identityadapter

import (
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func mapVenue(item *identityV1.TenantVenue) model.Venue {
	return model.Venue{
		ID:             item.GetId(),
		OrganizationID: item.GetOrganizationId(),
		Name:           item.GetName(),
		Slug:           item.GetSlug(),
		Address:        item.GetAddress(),
		Timezone:       item.GetTimezone(),
		Status:         item.GetStatus(),
	}
}

func mapVenues(items []*identityV1.TenantVenue) []model.Venue {
	result := make([]model.Venue, 0, len(items))
	for _, item := range items {
		result = append(result, mapVenue(item))
	}

	return result
}
