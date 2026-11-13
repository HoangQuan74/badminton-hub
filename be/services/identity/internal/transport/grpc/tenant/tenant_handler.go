package tenantgrpc

import (
	"context"
	"log/slog"

	grpcMetadata "github.com/kainguyen/badminton-hub/be/platform/grpc/metadata"
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/tenant"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *TenantServer) ListOrganizations(
	ctx context.Context,
	req *identityV1.TenantListOrganizationsRequest,
) (*identityV1.TenantOrganizationsResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Load organizations.
	rows, err := s.service.ListOrganizations(ctx, actor)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return organizationListResponse(rows), nil
}

func (s *TenantServer) GetOrganization(
	ctx context.Context,
	req *identityV1.TenantGetOrganizationRequest,
) (*identityV1.TenantOrganizationResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Load organization.
	item, err := s.service.GetOrganization(ctx, actor, req.GetOrganizationId())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantOrganizationResponse{Organization: organizationResponse(item)}, nil
}

func (s *TenantServer) UpdateOrganization(
	ctx context.Context,
	req *identityV1.TenantUpdateOrganizationRequest,
) (*identityV1.TenantOrganizationResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Update organization.
	input := tenant.UpdateOrganizationInput{
		Name: req.Name,
		Slug: req.Slug,
	}

	item, err := s.service.UpdateOrganization(ctx, actor, req.GetOrganizationId(), input)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantOrganizationResponse{Organization: organizationResponse(item)}, nil
}

func (s *TenantServer) ArchiveOrganization(
	ctx context.Context,
	req *identityV1.TenantArchiveOrganizationRequest,
) (*identityV1.TenantOrganizationResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.actor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Archive organization.
	item, err := s.service.ArchiveOrganization(ctx, actor, req.GetOrganizationId())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantOrganizationResponse{Organization: organizationResponse(item)}, nil
}

func (s *TenantServer) CreateOrganization(ctx context.Context, request *identityV1.TenantCreateOrganizationRequest) (*identityV1.TenantOrganizationResponse, error) {
	// Verify the caller before domain execution; no request actor ID is trusted.
	actorID, err := s.tokens.ParseAccessToken(request.GetAccessToken())
	if err != nil {
		return nil, identityException.TenantCreateGRPCMapper.Map(identityException.ErrAuthInvalidToken)
	}

	// Bind enrollment input after verifying the caller.
	input := tenant.CreateOrganizationInput{
		Name:           request.GetName(),
		Slug:           request.GetSlug(),
		IdempotencyKey: request.GetIdempotencyKey(),
		ContactName:    request.GetContactName(),
		ContactEmail:   request.GetContactEmail(),
		ContactPhone:   request.GetContactPhone(),
		Note:           request.GetNote(),
	}

	output, err := s.service.CreateOrganization(ctx, actorID, input)
	if err != nil {
		mappedError := identityException.TenantCreateGRPCMapper.Map(identityException.TenantCreateGRPCDetails(err))
		logCreateOrganizationFailure(ctx, mappedError)
		return nil, mappedError
	}

	return createResponse(output), nil
}

func logCreateOrganizationFailure(ctx context.Context, err error) {
	if status.Code(err) != codes.Internal {
		return
	}

	// Record correlation metadata without request payloads, credentials, or keys.
	metadata, _ := grpcMetadata.FromContext(ctx)
	slog.ErrorContext(ctx, "owner enrollment failed",
		"operation", "tenant.CreateOrganization",
		"request_id", metadata.RequestID,
		"correlation_id", metadata.CorrelationID,
	)
}

func (s *TenantServer) ListRegistrations(ctx context.Context, request *identityV1.TenantListRegistrationsRequest) (*identityV1.TenantRegistrationsResponse, error) {
	if _, err := s.authorizer.RequireAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	items, err := s.service.ListRegistrations(ctx, request.GetStatus())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return registrationListResponse(items), nil
}

func (s *TenantServer) ReviewRegistration(ctx context.Context, request *identityV1.TenantReviewRegistrationRequest) (*identityV1.TenantRegistrationResponse, error) {
	actor, err := s.authorizer.RequireAdmin(ctx, request.GetAccessToken())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	input := tenant.ReviewRegistrationInput{
		OrganizationID: request.GetOrganizationId(),
		Decision:       request.GetDecision(),
		Note:           request.GetNote(),
	}

	item, err := s.service.ReviewRegistration(ctx, actor.ID, input)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.TenantRegistrationResponse{Registration: registrationResponse(item)}, nil
}
