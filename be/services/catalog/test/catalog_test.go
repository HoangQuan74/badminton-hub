package test

import (
	"context"
	"testing"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

type repositoryStub struct {
	createCalls int
	name        string
	interval    int32
}

func (r *repositoryStub) ListOrganizations(context.Context, string) ([]model.Organization, error) {
	return nil, nil
}
func (r *repositoryStub) GetOrganization(context.Context, string, string) (model.Organization, error) {
	return model.Organization{}, nil
}
func (r *repositoryStub) CreateOrganization(context.Context, string, catalog.CreateOrganizationInput) (model.Organization, error) {
	return model.Organization{}, nil
}
func (r *repositoryStub) UpdateOrganization(context.Context, string, string, catalog.UpdateOrganizationInput) (model.Organization, error) {
	return model.Organization{}, nil
}
func (r *repositoryStub) ArchiveOrganization(context.Context, string, string) (model.Organization, error) {
	return model.Organization{}, nil
}

func (r *repositoryStub) GetPublicVenue(context.Context, string) (model.Venue, error) {
	return model.Venue{}, nil
}
func (r *repositoryStub) GetTenantVenue(context.Context, string, string) (model.Venue, error) {
	return model.Venue{}, nil
}
func (r *repositoryStub) CreateVenue(context.Context, string, catalog.CreateVenueInput) (model.Venue, error) {
	return model.Venue{}, nil
}
func (r *repositoryStub) UpdateVenue(context.Context, string, string, catalog.UpdateVenueInput) (model.Venue, error) {
	return model.Venue{}, nil
}
func (r *repositoryStub) ArchiveVenue(context.Context, string, string) (model.Venue, error) {
	return model.Venue{}, nil
}
func (r *repositoryStub) GetTenantCourt(context.Context, string, string) (model.Court, error) {
	return model.Court{}, nil
}

func (r *repositoryStub) ListPublicVenues(context.Context) ([]model.Venue, error) { return nil, nil }
func (r *repositoryStub) ListPublicVenueCourts(context.Context, string) ([]model.Court, error) {
	return nil, nil
}
func (r *repositoryStub) ListTenantVenues(context.Context, string) ([]model.Venue, error) {
	return nil, nil
}
func (r *repositoryStub) ListTenantCourts(context.Context, string, string) ([]model.Court, error) {
	return nil, nil
}
func (r *repositoryStub) CreateCourt(_ context.Context, _, _ string, name string, interval int32) (model.Court, error) {
	r.createCalls++
	r.name = name
	r.interval = interval
	return model.Court{Name: name, SlotIntervalMinutes: interval}, nil
}
func (r *repositoryStub) UpdateCourt(context.Context, string, string, string, int32, string) (model.Court, error) {
	return model.Court{}, nil
}
func (r *repositoryStub) ArchiveCourt(context.Context, string, string) (model.Court, error) {
	return model.Court{}, nil
}

func TestCreateCourtNormalizesNameAndDefaultsInterval(t *testing.T) {
	repository := &repositoryStub{}
	service := catalog.NewService(repository)

	court, err := service.CreateCourt(context.Background(), "actor", "venue", "  Sân 4  ", 0)
	if err != nil {
		t.Fatalf("CreateCourt() error = %v", err)
	}
	if court.Name != "Sân 4" || repository.name != "Sân 4" {
		t.Fatalf("name = %q, want trimmed Sân 4", court.Name)
	}
	if repository.interval != 30 || court.SlotIntervalMinutes != 30 {
		t.Fatalf("interval = %d, want 30", court.SlotIntervalMinutes)
	}
	if repository.createCalls != 1 {
		t.Fatalf("create calls = %d, want 1", repository.createCalls)
	}
}

func TestCreateCourtRejectsInvalidIntervalBeforeRepository(t *testing.T) {
	repository := &repositoryStub{}
	service := catalog.NewService(repository)

	_, err := service.CreateCourt(context.Background(), "actor", "venue", "Sân 4", 20)
	if err != catalog.ErrInvalidInput {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
	if repository.createCalls != 0 {
		t.Fatalf("create calls = %d, want 0", repository.createCalls)
	}
}
