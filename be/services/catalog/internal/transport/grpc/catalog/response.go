package cataloggrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	"time"
)

func mapVenues(values []model.Venue) []*catalogV1.Venue {
	items := make([]*catalogV1.Venue, 0, len(values))
	for _, value := range values {
		items = append(items, mapVenue(value))
	}
	return items
}

func mapCourts(values []model.Court) []*catalogV1.Court {
	items := make([]*catalogV1.Court, 0, len(values))
	for _, value := range values {
		items = append(items, mapCourt(value))
	}
	return items
}

func mapCourt(value model.Court) *catalogV1.Court {
	return &catalogV1.Court{
		Id:                  value.ID,
		VenueId:             value.VenueID,
		OrganizationId:      value.OrganizationID,
		Name:                value.Name,
		Status:              value.Status,
		SlotIntervalMinutes: value.SlotIntervalMinutes,
		CreatedAt:           value.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:           value.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func mapVenue(value model.Venue) *catalogV1.Venue {
	return &catalogV1.Venue{
		Id:             value.ID,
		OrganizationId: value.OrganizationID,
		Name:           value.Name,
		Slug:           value.Slug,
		Address:        value.Address,
		Timezone:       value.Timezone,
		Status:         value.Status,
		Courts:         mapCourts(value.Courts),
	}
}
