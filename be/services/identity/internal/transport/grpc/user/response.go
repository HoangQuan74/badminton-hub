package usergrpc

import (
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func userResponse(value model.User) *identityV1.UserResponse {
	// Map the domain user to the gRPC response.
	return &identityV1.UserResponse{
		Id:        value.ID,
		Email:     value.Email,
		FullName:  value.FullName,
		Phone:     value.Phone,
		Status:    string(value.Status),
		CreatedAt: value.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
}

func userListResponse(items []model.User) *identityV1.UserListResponse {
	// Map all domain users to the list response.
	response := &identityV1.UserListResponse{
		Users: make([]*identityV1.UserResponse, 0, len(items)),
	}
	for _, item := range items {
		response.Users = append(response.Users, userResponse(item))
	}
	return response
}
