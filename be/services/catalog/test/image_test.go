package test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

type imageRepositoryStub struct {
	authorized error
	saveErr    error
	saved      model.VenueImage
}

func (r *imageRepositoryStub) RequireVenueOwner(context.Context, string, string) error {
	return r.authorized
}
func (r *imageRepositoryStub) SaveVenueImage(_ context.Context, _ string, item model.VenueImage) (model.VenueImage, error) {
	r.saved = item
	return item, r.saveErr
}
func (*imageRepositoryStub) ListPublicVenueImages(context.Context, string) ([]model.VenueImage, error) {
	return nil, nil
}
func (*imageRepositoryStub) ListTenantVenueImages(context.Context, string, string) ([]model.VenueImage, error) {
	return nil, nil
}
func (*imageRepositoryStub) GetPublicVenueImage(context.Context, string) (model.VenueImage, error) {
	return model.VenueImage{}, nil
}
func (*imageRepositoryStub) GetTenantVenueImage(context.Context, string, string) (model.VenueImage, error) {
	return model.VenueImage{}, nil
}

type imageStoreStub struct {
	puts        int
	deletes     int
	contentType string
	key         string
	putErr      error
}

func (s *imageStoreStub) Put(_ context.Context, key, contentType string, _ []byte) error {
	s.puts++
	s.key = key
	s.contentType = contentType
	return s.putErr
}
func (*imageStoreStub) Get(context.Context, string, int64) ([]byte, error) { return nil, nil }
func (s *imageStoreStub) Delete(context.Context, string) error             { s.deletes++; return nil }

func samplePNG(t *testing.T) []byte {
	t.Helper()
	source := image.NewRGBA(image.Rect(0, 0, 1, 1))
	source.Set(0, 0, color.RGBA{R: 20, A: 255})
	var result bytes.Buffer
	if err := png.Encode(&result, source); err != nil {
		t.Fatal(err)
	}
	return result.Bytes()
}

func TestUploadVenueImageValidatesAndStores(t *testing.T) {
	repository := &imageRepositoryStub{}
	store := &imageStoreStub{}
	service := catalog.NewImageService(repository, store)
	item, err := service.Upload(context.Background(), "actor", "venue", samplePNG(t))
	if err != nil {
		t.Fatal(err)
	}
	if store.puts != 1 || store.contentType != "image/png" || item.ObjectKey == "" || repository.saved.ID != item.ID {
		t.Fatalf("store=%+v image=%+v", store, item)
	}
}

func TestUploadVenueImageRejectsInvalidAndUnauthorized(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		authErr error
		want    error
	}{
		{name: "empty", want: catalog.ErrInvalidInput},
		{name: "invalid file", content: []byte("not an image"), want: catalog.ErrInvalidInput},
		{name: "too large", content: make([]byte, catalog.MaxVenueImageBytes+1), want: catalog.ErrInvalidInput},
		{name: "staff forbidden", content: samplePNG(t), authErr: catalog.ErrForbidden, want: catalog.ErrForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &imageStoreStub{}
			service := catalog.NewImageService(&imageRepositoryStub{authorized: tt.authErr}, store)
			_, err := service.Upload(context.Background(), "actor", "venue", tt.content)
			if !errors.Is(err, tt.want) || store.puts != 0 {
				t.Fatalf("error=%v puts=%d", err, store.puts)
			}
		})
	}
}

func TestUploadVenueImageCleansUpAfterDatabaseFailure(t *testing.T) {
	store := &imageStoreStub{}
	service := catalog.NewImageService(&imageRepositoryStub{saveErr: errors.New("database unavailable")}, store)
	_, err := service.Upload(context.Background(), "actor", "venue", samplePNG(t))
	if err == nil || store.puts != 1 || store.deletes != 1 {
		t.Fatalf("error=%v puts=%d deletes=%d", err, store.puts, store.deletes)
	}
}

func TestUploadVenueImageReportsStorageFailure(t *testing.T) {
	store := &imageStoreStub{putErr: errors.New("storage offline")}
	repository := &imageRepositoryStub{}
	_, err := catalog.NewImageService(repository, store).Upload(context.Background(), "actor", "venue", samplePNG(t))
	if !errors.Is(err, catalog.ErrImageStorage) || repository.saved.ID != "" || store.deletes != 1 {
		t.Fatalf("error=%v saved=%+v", err, repository.saved)
	}
}
