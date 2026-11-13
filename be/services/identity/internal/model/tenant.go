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

type TenantRegistration struct {
	Organization                                  Organization
	ContactName, ContactEmail, ContactPhone, Note string
	ReviewDecision, ReviewNote, ReviewerID        string
	ReviewedAt                                    time.Time
}
