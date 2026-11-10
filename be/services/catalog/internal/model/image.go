package model

import "time"

// VenueImage contains persisted metadata; object keys are never exposed to clients.
type VenueImage struct {
	ID          string
	VenueID     string
	ObjectKey   string
	ContentType string
	SizeBytes   int32
	CreatedAt   time.Time
}
