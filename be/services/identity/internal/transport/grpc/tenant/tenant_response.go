package tenantgrpc

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func organizationResponse(item model.Organization) *identityV1.TenantOrganization {
	return &identityV1.TenantOrganization{
		Id:        item.ID,
		Name:      item.Name,
		Slug:      item.Slug,
		Status:    item.Status,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func createResponse(output tenant.CreateOrganizationOutput) *identityV1.TenantOrganizationResponse {
	// Include the enrolled role without exposing credentials or token material.
	user := grpcHelper.MapUserResponse(output.User)
	user.Roles = []string{output.User.RoleName}
	user.IsAdmin = output.User.IsAdmin
	return &identityV1.TenantOrganizationResponse{
		User:                    user,
		Organization:            organizationResponse(output.Organization),
		Membership:              grpcHelper.MapOrganizationStaffResponse(output.Membership),
		IdempotencyReplayed:     output.Replayed,
		IdempotencyKeyExpiresAt: output.ExpiresAt.UTC().Format(time.RFC3339Nano),
	}
}

func registrationResponse(item model.TenantRegistration) *identityV1.TenantRegistration {
	reviewedAt := ""
	if !item.ReviewedAt.IsZero() {
		reviewedAt = item.ReviewedAt.UTC().Format(time.RFC3339Nano)
	}

	return &identityV1.TenantRegistration{
		Organization:   organizationResponse(item.Organization),
		ContactName:    item.ContactName,
		ContactEmail:   item.ContactEmail,
		ContactPhone:   item.ContactPhone,
		Note:           item.Note,
		ReviewDecision: item.ReviewDecision,
		ReviewNote:     item.ReviewNote,
		ReviewerId:     item.ReviewerID,
		ReviewedAt:     reviewedAt,
	}
}

func registrationListResponse(items []model.TenantRegistration) *identityV1.TenantRegistrationsResponse {
	response := &identityV1.TenantRegistrationsResponse{Items: make([]*identityV1.TenantRegistration, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, registrationResponse(item))
	}

	return response
}

func organizationListResponse(rows []model.Organization) *identityV1.TenantOrganizationsResponse {
	items := make([]*identityV1.TenantOrganization, 0, len(rows))
	for _, row := range rows {
		items = append(items, organizationResponse(row))
	}

	return &identityV1.TenantOrganizationsResponse{Items: items}
}
