package venue

type CreateVenueInput struct {
	OrganizationID string `validate:"required"`
	Name           string `validate:"required"`
	Slug           string `validate:"required"`
	Address        string
	Timezone       string
}

type UpdateVenueInput struct{ Name, Slug, Address, Timezone *string }
