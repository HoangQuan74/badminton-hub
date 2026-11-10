package test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
)

type venueRepositoryStub struct {
	repositoryStub
	input catalog.CreateVenueInput
	calls int
}

func (r *venueRepositoryStub) CreateVenue(_ context.Context, _ string, input catalog.CreateVenueInput) (model.Venue, error) {
	r.input = input
	r.calls++
	return model.Venue{Name: input.Name, Slug: input.Slug, Timezone: input.Timezone}, nil
}

func TestCreateVenueNormalizesUnicodeAndDefaultsTimezone(t *testing.T) {
	repository := &venueRepositoryStub{}
	service := catalog.NewService(repository)
	item, err := service.CreateVenue(context.Background(), "actor", catalog.CreateVenueInput{OrganizationID: "org", Name: "  " + strings.Repeat("â", 100) + "  ", Slug: "  SAN-DEP  "})
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != strings.Repeat("â", 100) || item.Slug != "san-dep" || item.Timezone != "Asia/Ho_Chi_Minh" || repository.calls != 1 {
		t.Fatalf("venue=%+v calls=%d", item, repository.calls)
	}
}

func TestCreateVenueRejectsInvalidInput(t *testing.T) {
	tests := []catalog.CreateVenueInput{
		{OrganizationID: "org", Name: "  ", Slug: "valid"},
		{OrganizationID: "org", Name: "name", Slug: "not valid"},
		{OrganizationID: "org", Name: strings.Repeat("a", 121), Slug: "valid"},
		{OrganizationID: "org", Name: "name", Slug: "valid", Timezone: "Invalid/Timezone"},
	}
	for _, input := range tests {
		repository := &venueRepositoryStub{}
		_, err := catalog.NewService(repository).CreateVenue(context.Background(), "actor", input)
		if !errors.Is(err, catalog.ErrInvalidInput) || repository.calls != 0 {
			t.Fatalf("input=%+v error=%v calls=%d", input, err, repository.calls)
		}
	}
}
