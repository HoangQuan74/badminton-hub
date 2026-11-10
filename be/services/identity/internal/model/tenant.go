package model

import "time"

type Organization struct {
	ID        string
	Name      string
	Slug      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Venue struct {
	ID             string
	OrganizationID string
	Name           string
	Slug           string
	Address        string
	Timezone       string
	Status         string
}

type OrganizationStaff struct{ OrganizationID, UserID, Role, Status string }
type VenueStaff struct{ OrganizationID, VenueID, UserID, Status string }
