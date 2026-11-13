package test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	catalogpostgres "github.com/kainguyen/badminton-hub/be/services/catalog/internal/repository/postgres/catalog"
)

// Use only a disposable database with Catalog migrations and development seed.
func TestCatalogPostgresIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_CATALOG_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_CATALOG_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repository := catalogpostgres.NewRepository(pool)
	organization := "10000000-0000-4000-8000-000000000001"
	venue := "20000000-0000-4000-8000-000000000001"
	courts, err := repository.ListCourts(ctx, venue, organization, true)
	if err != nil || len(courts) != 3 {
		t.Fatalf("seeded courts = %#v, error = %v", courts, err)
	}
	if _, err := repository.CreateCourt(ctx, organization, venue, "Sân 1", 30); !errors.Is(err, catalogException.ErrCourtNameConflict) {
		t.Fatalf("duplicate error = %v", err)
	}
	court, err := repository.CreateCourt(ctx, organization, venue, "Integration "+uuid.NewString(), 30)
	if err != nil {
		t.Fatal(err)
	}
	found, err := repository.GetCourt(ctx, court.ID)
	if err != nil || found.VenueID != venue || found.OrganizationID != organization {
		t.Fatalf("court = %#v, error = %v", found, err)
	}
	updatedName := "Updated " + uuid.NewString()
	updated, err := repository.UpdateCourt(ctx, court, updatedName, 60, "ACTIVE")
	if err != nil || updated.Name != updatedName || updated.SlotIntervalMinutes != 60 {
		t.Fatalf("updated = %#v, error = %v", updated, err)
	}
	archived, err := repository.ArchiveCourt(ctx, updated)
	if err != nil || archived.Status != "INACTIVE" {
		t.Fatalf("archived = %#v, error = %v", archived, err)
	}
	pool.Close()
	if _, err := repository.GetCourt(ctx, court.ID); err == nil {
		t.Fatal("closed database returned success")
	}
}
