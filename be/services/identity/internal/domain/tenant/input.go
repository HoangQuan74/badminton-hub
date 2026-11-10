package tenant

type CreateOrganizationInput struct {
	Name string `validate:"required"`
	Slug string `validate:"required"`
}

type UpdateOrganizationInput struct {
	Name *string
	Slug *string
}

type CreateVenueInput struct {
	OrganizationID string `validate:"required"`
	Name           string `validate:"required"`
	Slug           string `validate:"required"`
	Address        string
	Timezone       string
}

type UpdateVenueInput struct{ Name, Slug, Address, Timezone *string }
type SetOrganizationStaffInput struct{ OrganizationID, UserID, Role, Status string }
type SetVenueStaffInput struct{ VenueID, UserID, Status string }
