package usergrpc

import (
	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func userListResponse(items []model.User) *identityV1.UserListResponse {
	// Map all domain users to the list response.
	response := &identityV1.UserListResponse{
		Users: make([]*identityV1.UserResponse, 0, len(items)),
	}
	for _, item := range items {
		response.Users = append(response.Users, grpcHelper.MapUserResponse(item))
	}

	return response
}
