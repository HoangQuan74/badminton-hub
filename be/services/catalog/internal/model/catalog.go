package model

import "time"

type Venue struct {
	ID             string
	OrganizationID string
	Name           string
	Slug           string
	Address        string
	Timezone       string
	Status         string
	Courts         []Court
}

type Court struct {
	ID                  string
	VenueID             string
	OrganizationID      string
	Name                string
	Status              string
	SlotIntervalMinutes int32
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
