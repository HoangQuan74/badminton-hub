package cataloggrpc

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
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

func mapOrganization(item model.Organization) *catalogV1.Organization {
	return &catalogV1.Organization{
		Id:        item.ID,
		Name:      item.Name,
		Slug:      item.Slug,
		Status:    item.Status,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func mapImages(items []model.VenueImage) []*catalogV1.VenueImage {
	result := make([]*catalogV1.VenueImage, 0, len(items))
	for _, item := range items {
		result = append(result, mapImage(item))
	}
	return result
}

func mapImage(item model.VenueImage) *catalogV1.VenueImage {
	return &catalogV1.VenueImage{
		Id:          item.ID,
		VenueId:     item.VenueID,
		ContentType: item.ContentType,
		SizeBytes:   item.SizeBytes,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}
