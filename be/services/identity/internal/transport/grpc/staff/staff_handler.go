package staffgrpc

import (
	"context"

	staffDomain "github.com/kainguyen/badminton-hub/be/services/identity/internal/domain/staff"
	identityException "github.com/kainguyen/badminton-hub/be/services/identity/internal/exception"
	grpcHelper "github.com/kainguyen/badminton-hub/be/services/identity/internal/transport/grpc/helper"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func (s *StaffServer) ListStaff(ctx context.Context, request *identityV1.ListStaffRequest) (*identityV1.StaffListResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	// Build the shared pagination and filter query.
	query, err := grpcHelper.ParseListQuery(request)
	if err != nil {
		return nil, identityException.StaffGRPCMapper.Map(identityException.ErrStaffInvalidInput)
	}

	// Load the requested staff members from the domain service.
	items, err := s.service.ListStaff(ctx, query)
	if err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	return staffListResponse(items), nil
}

func (s *StaffServer) CreateStaff(ctx context.Context, request *identityV1.CreateStaffRequest) (*identityV1.StaffResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	// Map the gRPC request to the domain input.
	input := staffDomain.CreateInput{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
		FullName: request.GetFullName(),
		IsAdmin:  request.GetIsAdmin(),
		RoleID:   request.GetRoleId(),
	}

	// Create the staff member in the domain service.
	item, err := s.service.CreateStaff(ctx, input)
	if err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	return staffResponse(item), nil
}

func (s *StaffServer) GetStaff(ctx context.Context, request *identityV1.GetStaffRequest) (*identityV1.StaffResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	// Load the requested staff member from the domain service.
	item, err := s.service.GetStaff(ctx, request.GetStaffId())
	if err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	return staffResponse(item), nil
}

func (s *StaffServer) UpdateStaff(ctx context.Context, request *identityV1.UpdateStaffRequest) (*identityV1.StaffResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	// Map the gRPC request to the domain input.
	input := staffDomain.UpdateStaffInput{
		ID:       request.GetStaffId(),
		Email:    request.GetEmail(),
		FullName: request.GetFullName(),
		Status:   request.GetStatus(),
	}

	// Update the staff member in the domain service.
	item, err := s.service.UpdateStaff(ctx, input)
	if err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	return staffResponse(item), nil
}

func (s *StaffServer) DeleteStaff(ctx context.Context, request *identityV1.DeleteStaffRequest) (*identityV1.Empty, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	// Delete the staff member in the domain service.
	if err := s.service.DeleteStaff(ctx, request.GetStaffId()); err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	return &identityV1.Empty{}, nil
}

func (s *StaffServer) UpdateStaffAccess(
	ctx context.Context,
	request *identityV1.UpdateStaffAccessRequest,
) (*identityV1.StaffResponse, error) {
	// Authorize admin access.
	if err := s.authorizer.AuthorizeAdmin(ctx, request.GetAccessToken()); err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	// Map the gRPC request to the domain input.
	input := staffDomain.UpdateStaffAccessInput{
		ID:      request.GetStaffId(),
		IsAdmin: request.GetIsAdmin(),
		RoleID:  request.GetRoleId(),
	}

	// Update access in the domain service.
	item, err := s.service.UpdateStaffAccess(ctx, input)
	if err != nil {
		return nil, identityException.StaffGRPCMapper.Map(err)
	}

	return staffResponse(item), nil
}

func (s *StaffServer) SetOrganizationStaff(
	ctx context.Context,
	req *identityV1.TenantSetOrganizationStaffRequest,
) (*identityV1.TenantStaffResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.tenantActor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Update organization staff.
	input := staffDomain.SetOrganizationStaffInput{
		OrganizationID: req.GetOrganizationId(),
		UserID:         req.GetUserId(),
		Role:           req.GetRole(),
		Status:         req.GetStatus(),
	}

	item, err := s.memberships.SetOrganizationStaff(ctx, actor, input)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return grpcHelper.MapOrganizationStaffResponse(item), nil
}

func (s *StaffServer) SetVenueStaff(
	ctx context.Context,
	req *identityV1.TenantSetVenueStaffRequest,
) (*identityV1.TenantVenueStaffResponse, error) {
	// Verify the authenticated actor.
	actor, err := s.tenantActor(ctx, req.GetAccessToken(), false)
	if err != nil {
		return nil, err
	}

	// Update venue staff.
	input := staffDomain.SetVenueStaffInput{
		VenueID: req.GetVenueId(),
		UserID:  req.GetUserId(),
		Status:  req.GetStatus(),
	}

	item, err := s.memberships.SetVenueStaff(ctx, actor, input)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return venueStaffResponse(item), nil
}

func (s *StaffServer) ListOrganizationStaff(ctx context.Context, request *identityV1.ListOrganizationStaffRequest) (*identityV1.OrganizationStaffListResponse, error) {
	actor, err := s.tenantActor(ctx, request.GetAccessToken(), true)
	if err != nil {
		return nil, err
	}

	items, err := s.memberships.ListOrganizationStaff(ctx, actor, request.GetOrganizationId())
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return organizationStaffListResponse(items), nil
}

func (s *StaffServer) AddOrganizationStaff(ctx context.Context, request *identityV1.AddOrganizationStaffRequest) (*identityV1.TenantStaffResponse, error) {
	actor, err := s.tenantActor(ctx, request.GetAccessToken(), true)
	if err != nil {
		return nil, err
	}

	input := staffDomain.AddOrganizationStaffInput{
		OrganizationID: request.GetOrganizationId(),
		Email:          request.GetEmail(),
		Role:           request.GetRole(),
	}

	item, err := s.memberships.AddOrganizationStaff(ctx, actor, input)
	if err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return grpcHelper.MapOrganizationStaffResponse(item), nil
}

func (s *StaffServer) RemoveOrganizationStaff(ctx context.Context, request *identityV1.RemoveOrganizationStaffRequest) (*identityV1.Empty, error) {
	actor, err := s.tenantActor(ctx, request.GetAccessToken(), true)
	if err != nil {
		return nil, err
	}

	input := staffDomain.RemoveOrganizationStaffInput{
		OrganizationID: request.GetOrganizationId(),
		UserID:         request.GetUserId(),
	}

	if err := s.memberships.RemoveOrganizationStaff(ctx, actor, input); err != nil {
		return nil, identityException.TenantGRPCMapper.Map(err)
	}

	return &identityV1.Empty{}, nil
}
