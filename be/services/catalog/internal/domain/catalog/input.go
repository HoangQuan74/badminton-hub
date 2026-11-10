package catalog

// courtInput holds the values validated before canonicalization and business rules.
type courtInput struct {
	Name                string `validate:"required"`
	SlotIntervalMinutes int32
	Status              string
}

// CreateVenueInput is scoped to an organization owned by the authenticated actor.
type CreateVenueInput struct {
	OrganizationID string `validate:"required"`
	Name           string `validate:"required"`
	Slug           string `validate:"required"`
	Address        string
	Timezone       string
}

// UpdateVenueInput contains only fields supplied by the caller.
type UpdateVenueInput struct {
	Name     *string
	Slug     *string
	Address  *string
	Timezone *string
}

type CreateOrganizationInput struct {
	Name string `validate:"required"`
	Slug string `validate:"required"`
}

type UpdateOrganizationInput struct {
	Name *string
	Slug *string
}
