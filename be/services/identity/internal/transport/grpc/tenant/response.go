package tenantgrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"time"
)

func organizationResponse(item model.Organization) *identityV1.TenantOrganization {
	return &identityV1.TenantOrganization{Id: item.ID, Name: item.Name, Slug: item.Slug, Status: item.Status, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func venueResponse(item model.Venue) *identityV1.TenantVenue {
	return &identityV1.TenantVenue{Id: item.ID, OrganizationId: item.OrganizationID, Name: item.Name, Slug: item.Slug, Address: item.Address, Timezone: item.Timezone, Status: item.Status}
}
