package model

import "time"

type Staff struct {
	ID        string
	Email     string
	FullName  string
	IsAdmin   bool
	RoleID    string
	RoleName  string
	Status    string
	CreatedAt time.Time
}

type OrganizationStaff struct{ OrganizationID, UserID, Role, Status string }
type VenueStaff struct{ OrganizationID, VenueID, UserID, Status string }

type OrganizationStaffDetails struct {
	OrganizationStaff
	Email, FullName string
	VenueIDs        []string
}
