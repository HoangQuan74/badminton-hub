package catalog

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

const MaxVenueImageBytes = 2 * 1024 * 1024
const maxImagePixels = 16_000_000

type ImageRepository interface {
	RequireVenueOwner(context.Context, string, string) error
	SaveVenueImage(context.Context, string, model.VenueImage) (model.VenueImage, error)
	ListPublicVenueImages(context.Context, string) ([]model.VenueImage, error)
	ListTenantVenueImages(context.Context, string, string) ([]model.VenueImage, error)
	GetPublicVenueImage(context.Context, string) (model.VenueImage, error)
	GetTenantVenueImage(context.Context, string, string) (model.VenueImage, error)
}

type ImageStore interface {
	Put(context.Context, string, string, []byte) error
	Get(context.Context, string, int64) ([]byte, error)
	Delete(context.Context, string) error
}

type ImageService struct {
	repository ImageRepository
	store      ImageStore
}

func NewImageService(repository ImageRepository, store ImageStore) *ImageService {
	return &ImageService{repository: repository, store: store}
}

func (s *ImageService) Upload(ctx context.Context, actorID, venueID string, content []byte) (model.VenueImage, error) {
	if err := s.repository.RequireVenueOwner(ctx, actorID, venueID); err != nil {
		return model.VenueImage{}, err
	}
	contentType, extension, err := validateImage(content)
	if err != nil {
		return model.VenueImage{}, err
	}
	// Use server-generated keys; no user-provided filename reaches object storage.
	imageID := uuid.NewString()
	item := model.VenueImage{ID: imageID, VenueID: venueID, ObjectKey: "venues/" + venueID + "/" + imageID + extension, ContentType: contentType, SizeBytes: int32(len(content))}
	storeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := s.store.Put(storeCtx, item.ObjectKey, item.ContentType, content); err != nil {
		s.cleanup(item)
		return model.VenueImage{}, fmt.Errorf("%w: store venue image: %v", ErrImageStorage, err)
	}
	saved, err := s.repository.SaveVenueImage(ctx, actorID, item)
	if err != nil {
		s.cleanup(item)
		return model.VenueImage{}, err
	}
	return saved, nil
}

func (s *ImageService) ListPublic(ctx context.Context, venueID string) ([]model.VenueImage, error) {
	return s.repository.ListPublicVenueImages(ctx, venueID)
}
func (s *ImageService) ListTenant(ctx context.Context, actorID, venueID string) ([]model.VenueImage, error) {
	return s.repository.ListTenantVenueImages(ctx, actorID, venueID)
}
func (s *ImageService) GetPublic(ctx context.Context, imageID string) (model.VenueImage, []byte, error) {
	item, err := s.repository.GetPublicVenueImage(ctx, imageID)
	if err != nil {
		return model.VenueImage{}, nil, err
	}
	return s.get(ctx, item)
}
func (s *ImageService) GetTenant(ctx context.Context, actorID, imageID string) (model.VenueImage, []byte, error) {
	item, err := s.repository.GetTenantVenueImage(ctx, actorID, imageID)
	if err != nil {
		return model.VenueImage{}, nil, err
	}
	return s.get(ctx, item)
}
func (s *ImageService) get(ctx context.Context, item model.VenueImage) (model.VenueImage, []byte, error) {
	storeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	content, err := s.store.Get(storeCtx, item.ObjectKey, MaxVenueImageBytes)
	if err != nil {
		return model.VenueImage{}, nil, fmt.Errorf("%w: read venue image: %v", ErrImageStorage, err)
	}
	return item, content, nil
}

func validateImage(content []byte) (string, string, error) {
	if len(content) == 0 || len(content) > MaxVenueImageBytes {
		return "", "", ErrInvalidInput
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > maxImagePixels {
		return "", "", ErrInvalidInput
	}
	switch format {
	case "jpeg", "png":
		if _, _, err := image.Decode(bytes.NewReader(content)); err != nil {
			return "", "", ErrInvalidInput
		}
	default:
		return "", "", ErrInvalidInput
	}
	if format == "jpeg" {
		return "image/jpeg", ".jpg", nil
	}
	return "image/png", ".png", nil
}

func (s *ImageService) cleanup(item model.VenueImage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.store.Delete(ctx, item.ObjectKey); err != nil {
		slog.Error("venue image cleanup failed", "image_id", item.ID, "error", err)
	}
}
