package authgrpc

import (
	"time"

	model "github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func currentAdminStaffResponse(value model.Staff) *identityV1.StaffResponse {
	response := grpcHelper.MapStaffResponse(value)
	response.CreatedAt = value.CreatedAt.UTC().Format(time.RFC3339Nano)
	return response
}

func adminTokens(tokens model.TokenPair) *identityV1.TokenPair {
	return mapTokenPair(tokens, time.RFC3339)
}

func adminAuthResponse(staff model.Staff, tokens model.TokenPair) *identityV1.AdminAuthResponse {
	return &identityV1.AdminAuthResponse{
		Staff:  grpcHelper.MapStaffResponse(staff),
		Tokens: adminTokens(tokens),
	}
}
