package model

type Venue struct {
	ID             string
	OrganizationID string
	Name           string
	Slug           string
	Address        string
	Timezone       string
	Status         string
}

type VenueOrganizationAccess struct {
	Status string
	Role   string
}
