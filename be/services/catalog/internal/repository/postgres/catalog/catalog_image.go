package catalogpostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	db "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/postgres/generated/sqlc"
)

func (r *Repository) SaveImage(ctx context.Context, image model.VenueImage) (model.VenueImage, error) {
	id, err := parseID(image.ID)
	if err != nil {
		return model.VenueImage{}, err
	}

	venue, err := parseID(image.VenueID)
	if err != nil {
		return model.VenueImage{}, err
	}

	row, err := r.queries.CreateVenueImage(ctx, db.CreateVenueImageParams{
		ID:          id,
		VenueID:     venue,
		ObjectKey:   image.ObjectKey,
		ContentType: image.ContentType,
		SizeBytes:   image.SizeBytes,
	})
	if err != nil {
		return model.VenueImage{}, fmt.Errorf("save venue image: %w", err)
	}

	return imageModel(row), nil
}

func (r *Repository) ListImages(ctx context.Context, venueID string) ([]model.VenueImage, error) {
	id, err := parseID(venueID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListVenueImages(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list venue images: %w", err)
	}

	items := make(
		[]model.VenueImage,
		0,
		len(rows),
	)
	for _, row := range rows {
		items = append(items, imageModel(row))
	}

	return items, nil
}

func (r *Repository) GetImage(ctx context.Context, imageID string) (model.VenueImage, error) {
	id, err := parseID(imageID)
	if err != nil {
		return model.VenueImage{}, err
	}

	row, err := r.queries.GetVenueImage(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.VenueImage{}, catalogException.ErrNotFound
	}

	if err != nil {
		return model.VenueImage{}, fmt.Errorf("get venue image: %w", err)
	}

	return imageModel(row), nil
}

func imageModel(row db.CatalogVenueImage) model.VenueImage {
	return model.VenueImage{
		ID:          row.ID.String(),
		VenueID:     row.VenueID.String(),
		ObjectKey:   row.ObjectKey,
		ContentType: row.ContentType,
		SizeBytes:   row.SizeBytes,
		CreatedAt:   row.CreatedAt.Time,
	}
}
