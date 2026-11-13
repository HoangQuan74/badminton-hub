package cataloggrpc

import (
	"context"

	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

func (s *CatalogServer) UploadVenueImage(
	ctx context.Context,
	request *catalogV1.UploadVenueImageRequest,
) (*catalogV1.VenueImageResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	if s.images == nil {
		return nil, catalogException.ErrGRPCImageStorageUnavailable
	}

	item, err := s.images.Upload(ctx, actorID, request.GetVenueId(), request.GetContent())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueImageResponse{Image: mapImage(item)}, nil
}

func (s *CatalogServer) ListPublicVenueImages(
	ctx context.Context,
	request *catalogV1.ListPublicVenueImagesRequest,
) (*catalogV1.VenueImageListResponse, error) {
	if s.images == nil {
		return nil, catalogException.ErrGRPCImageStorageUnavailable
	}

	items, err := s.images.ListPublic(ctx, request.GetVenueId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueImageListResponse{Items: mapImages(items)}, nil
}

func (s *CatalogServer) ListTenantVenueImages(
	ctx context.Context,
	request *catalogV1.ListTenantVenueImagesRequest,
) (*catalogV1.VenueImageListResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	if s.images == nil {
		return nil, catalogException.ErrGRPCImageStorageUnavailable
	}

	items, err := s.images.ListTenant(ctx, actorID, request.GetVenueId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueImageListResponse{Items: mapImages(items)}, nil
}

func (s *CatalogServer) GetPublicVenueImage(
	ctx context.Context,
	request *catalogV1.GetVenueImageRequest,
) (*catalogV1.VenueImageContentResponse, error) {
	if s.images == nil {
		return nil, catalogException.ErrGRPCImageStorageUnavailable
	}

	item, content, err := s.images.GetPublic(ctx, request.GetImageId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueImageContentResponse{ContentType: item.ContentType, Content: content}, nil
}

func (s *CatalogServer) GetTenantVenueImage(
	ctx context.Context,
	request *catalogV1.GetVenueImageRequest,
) (*catalogV1.VenueImageContentResponse, error) {
	// Verify the authenticated actor.
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}

	if s.images == nil {
		return nil, catalogException.ErrGRPCImageStorageUnavailable
	}

	item, content, err := s.images.GetTenant(ctx, actorID, request.GetImageId())
	if err != nil {
		return nil, catalogException.GRPCMapper.Map(err)
	}

	return &catalogV1.VenueImageContentResponse{ContentType: item.ContentType, Content: content}, nil
}
