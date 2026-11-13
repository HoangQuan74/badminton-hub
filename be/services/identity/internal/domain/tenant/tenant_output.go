package tenant

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

// CreateOrganizationOutput is also the replay snapshot. Credentials are never populated.
type CreateOrganizationOutput struct {
	User         model.User              `json:"user"`
	Organization model.Organization      `json:"organization"`
	Membership   model.OrganizationStaff `json:"membership"`
	Replayed     bool                    `json:"-"`
	ExpiresAt    time.Time               `json:"-"`
}
