package helper

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

// MapUserResponse maps the public profile. Callers add access fields when required.
func MapUserResponse(user model.User) *identityV1.UserResponse {
	return &identityV1.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Phone:     user.Phone,
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}
