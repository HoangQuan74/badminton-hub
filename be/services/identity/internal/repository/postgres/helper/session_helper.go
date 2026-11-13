package helper

import (
	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/identity/internal/storage/postgres/generated/sqlc"
)

// MapSessionParams maps already parsed identifiers and the hashed session material.
func MapSessionParams(sessionID, userID uuid.UUID, session model.Session) db.CreateAuthSessionParams {
	return db.CreateAuthSessionParams{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: session.TokenHash,
		ExpiresAt:        MapTimestamp(session.ExpiresAt),
	}
}
