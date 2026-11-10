package cataloggrpc

import (
	"context"
	"time"

	grpcEndpoint "github.com/kainguyen/badminton-hub/be/platform/grpc/endpoint"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UploadVenueImage(ctx context.Context, request *catalogV1.UploadVenueImageRequest) (*catalogV1.VenueImageResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	if s.images == nil {
		return nil, status.Error(codes.Unavailable, "image storage unavailable")
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.UploadVenueImageRequest) (*catalogV1.VenueImageResponse, error) {
		item, err := s.images.Upload(ctx, actorID, request.GetVenueId(), request.GetContent())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueImageResponse{Image: mapImage(item)}, nil
	})
}

func (s *Server) ListPublicVenueImages(ctx context.Context, request *catalogV1.ListPublicVenueImagesRequest) (*catalogV1.VenueImageListResponse, error) {
	if s.images == nil {
		return nil, status.Error(codes.Unavailable, "image storage unavailable")
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ListPublicVenueImagesRequest) (*catalogV1.VenueImageListResponse, error) {
		items, err := s.images.ListPublic(ctx, request.GetVenueId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueImageListResponse{Items: mapImages(items)}, nil
	})
}

func (s *Server) ListTenantVenueImages(ctx context.Context, request *catalogV1.ListTenantVenueImagesRequest) (*catalogV1.VenueImageListResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	if s.images == nil {
		return nil, status.Error(codes.Unavailable, "image storage unavailable")
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.ListTenantVenueImagesRequest) (*catalogV1.VenueImageListResponse, error) {
		items, err := s.images.ListTenant(ctx, actorID, request.GetVenueId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueImageListResponse{Items: mapImages(items)}, nil
	})
}

func (s *Server) GetPublicVenueImage(ctx context.Context, request *catalogV1.GetVenueImageRequest) (*catalogV1.VenueImageContentResponse, error) {
	if s.images == nil {
		return nil, status.Error(codes.Unavailable, "image storage unavailable")
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.GetVenueImageRequest) (*catalogV1.VenueImageContentResponse, error) {
		item, content, err := s.images.GetPublic(ctx, request.GetImageId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueImageContentResponse{ContentType: item.ContentType, Content: content}, nil
	})
}

func (s *Server) GetTenantVenueImage(ctx context.Context, request *catalogV1.GetVenueImageRequest) (*catalogV1.VenueImageContentResponse, error) {
	actorID, err := s.tenantActorID(ctx)
	if err != nil {
		return nil, err
	}
	if s.images == nil {
		return nil, status.Error(codes.Unavailable, "image storage unavailable")
	}
	return grpcEndpoint.Run(ctx, request, errorMapper, nil, func(ctx context.Context, request *catalogV1.GetVenueImageRequest) (*catalogV1.VenueImageContentResponse, error) {
		item, content, err := s.images.GetTenant(ctx, actorID, request.GetImageId())
		if err != nil {
			return nil, err
		}
		return &catalogV1.VenueImageContentResponse{ContentType: item.ContentType, Content: content}, nil
	})
}

func mapImages(items []model.VenueImage) []*catalogV1.VenueImage {
	result := make([]*catalogV1.VenueImage, 0, len(items))
	for _, item := range items {
		result = append(result, mapImage(item))
	}
	return result
}

func mapImage(item model.VenueImage) *catalogV1.VenueImage {
	return &catalogV1.VenueImage{Id: item.ID, VenueId: item.VenueID, ContentType: item.ContentType, SizeBytes: item.SizeBytes, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano)}
}
