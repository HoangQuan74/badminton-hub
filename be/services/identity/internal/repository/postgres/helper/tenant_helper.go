package helper

import (
	"github.com/google/uuid"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
)

// ParseTenantScope keeps actor and resource parse errors consistent across adapters.
func ParseTenantScope(actorID, resourceID string) (uuid.UUID, uuid.UUID, error) {
	actor, err := uuid.Parse(actorID)
	if err != nil {
		return uuid.Nil, uuid.Nil, identityException.ErrTenantForbidden
	}

	resource, err := uuid.Parse(resourceID)
	if err != nil {
		return uuid.Nil, uuid.Nil, identityException.ErrTenantInvalidInput
	}

	return actor, resource, nil
}
