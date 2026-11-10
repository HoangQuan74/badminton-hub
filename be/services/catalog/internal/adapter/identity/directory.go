package identityadapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Directory struct {
	client identityV1.IdentityTenantServiceClient
}

func New(client identityV1.IdentityTenantServiceClient) *Directory { return &Directory{client: client} }

func token(ctx context.Context) (string, error) {
	incoming, _ := metadata.FromIncomingContext(ctx)
	values := incoming.Get("authorization")
	if len(values) != 1 {
		return "", catalog.ErrUnauthenticated
	}
	scheme, value, ok := strings.Cut(values[0], " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(value) == "" {
		return "", catalog.ErrUnauthenticated
	}
	return strings.TrimSpace(value), nil
}

func call[T any](ctx context.Context, operation func(context.Context) (T, error), conflict error) (T, error) {
	deadline, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	value, err := operation(deadline)
	if err == nil {
		return value, nil
	}
	var zero T
	switch status.Code(err) {
	case codes.Unauthenticated:
		return zero, catalog.ErrUnauthenticated
	case codes.PermissionDenied:
		return zero, catalog.ErrForbidden
	case codes.InvalidArgument:
		return zero, catalog.ErrInvalidInput
	case codes.NotFound:
		return zero, catalog.ErrNotFound
	case codes.AlreadyExists:
		if conflict != nil {
			return zero, conflict
		}
	case codes.FailedPrecondition:
		return zero, catalog.ErrOrganizationStateConflict
	case codes.Unavailable, codes.DeadlineExceeded:
		return zero, fmt.Errorf("%w: %v", catalog.ErrIdentityUnavailable, err)
	}
	return zero, fmt.Errorf("identity tenant request: %w", err)
}

func organization(item *identityV1.TenantOrganization) model.Organization {
	created, _ := time.Parse(time.RFC3339Nano, item.GetCreatedAt())
	updated, _ := time.Parse(time.RFC3339Nano, item.GetUpdatedAt())
	return model.Organization{ID: item.GetId(), Name: item.GetName(), Slug: item.GetSlug(), Status: item.GetStatus(), CreatedAt: created, UpdatedAt: updated}
}
func venue(item *identityV1.TenantVenue) model.Venue {
	return model.Venue{ID: item.GetId(), OrganizationID: item.GetOrganizationId(), Name: item.GetName(), Slug: item.GetSlug(), Address: item.GetAddress(), Timezone: item.GetTimezone(), Status: item.GetStatus()}
}
func venues(items []*identityV1.TenantVenue) []model.Venue {
	result := make([]model.Venue, 0, len(items))
	for _, item := range items {
		result = append(result, venue(item))
	}
	return result
}

func (d *Directory) ListOrganizations(ctx context.Context) ([]model.Organization, error) {
	access, err := token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantOrganizationsResponse, error) {
		return d.client.ListOrganizations(ctx, &identityV1.TenantListOrganizationsRequest{AccessToken: access})
	}, nil)
	if err != nil {
		return nil, err
	}
	items := make([]model.Organization, 0, len(result.GetItems()))
	for _, item := range result.GetItems() {
		items = append(items, organization(item))
	}
	return items, nil
}
func (d *Directory) GetOrganization(ctx context.Context, id string) (model.Organization, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Organization{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantOrganizationResponse, error) {
		return d.client.GetOrganization(ctx, &identityV1.TenantGetOrganizationRequest{AccessToken: access, OrganizationId: id})
	}, nil)
	if err != nil {
		return model.Organization{}, err
	}
	return organization(result.GetOrganization()), nil
}
func (d *Directory) CreateOrganization(ctx context.Context, input catalog.CreateOrganizationInput) (model.Organization, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Organization{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantOrganizationResponse, error) {
		return d.client.CreateOrganization(ctx, &identityV1.TenantCreateOrganizationRequest{AccessToken: access, Name: input.Name, Slug: input.Slug})
	}, catalog.ErrOrganizationSlugConflict)
	if err != nil {
		return model.Organization{}, err
	}
	return organization(result.GetOrganization()), nil
}
func (d *Directory) UpdateOrganization(ctx context.Context, id string, input catalog.UpdateOrganizationInput) (model.Organization, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Organization{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantOrganizationResponse, error) {
		return d.client.UpdateOrganization(ctx, &identityV1.TenantUpdateOrganizationRequest{AccessToken: access, OrganizationId: id, Name: input.Name, Slug: input.Slug})
	}, catalog.ErrOrganizationSlugConflict)
	if err != nil {
		return model.Organization{}, err
	}
	return organization(result.GetOrganization()), nil
}
func (d *Directory) ArchiveOrganization(ctx context.Context, id string) (model.Organization, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Organization{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantOrganizationResponse, error) {
		return d.client.ArchiveOrganization(ctx, &identityV1.TenantArchiveOrganizationRequest{AccessToken: access, OrganizationId: id})
	}, nil)
	if err != nil {
		return model.Organization{}, err
	}
	return organization(result.GetOrganization()), nil
}
func (d *Directory) ListPublicVenues(ctx context.Context) ([]model.Venue, error) {
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenuesResponse, error) {
		return d.client.ListPublicVenues(ctx, &identityV1.TenantListPublicVenuesRequest{})
	}, nil)
	if err != nil {
		return nil, err
	}
	return venues(result.GetItems()), nil
}
func (d *Directory) GetPublicVenue(ctx context.Context, slug string) (model.Venue, error) {
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenueResponse, error) {
		return d.client.GetPublicVenue(ctx, &identityV1.TenantGetPublicVenueRequest{Slug: slug})
	}, nil)
	if err != nil {
		return model.Venue{}, err
	}
	return venue(result.GetVenue()), nil
}
func (d *Directory) GetPublicVenueByID(ctx context.Context, id string) (model.Venue, error) {
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenueResponse, error) {
		return d.client.GetPublicVenueByID(ctx, &identityV1.TenantGetPublicVenueByIDRequest{VenueId: id})
	}, nil)
	if err != nil {
		return model.Venue{}, err
	}
	return venue(result.GetVenue()), nil
}
func (d *Directory) ListVenues(ctx context.Context) ([]model.Venue, error) {
	access, err := token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenuesResponse, error) {
		return d.client.ListVenues(ctx, &identityV1.TenantListVenuesRequest{AccessToken: access})
	}, nil)
	if err != nil {
		return nil, err
	}
	return venues(result.GetItems()), nil
}
func (d *Directory) AuthorizeVenue(ctx context.Context, id string, requireOwner bool) (model.Venue, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Venue{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenueResponse, error) {
		return d.client.AuthorizeVenue(ctx, &identityV1.TenantAuthorizeVenueRequest{AccessToken: access, VenueId: id, RequireOwner: requireOwner})
	}, nil)
	if err != nil {
		return model.Venue{}, err
	}
	return venue(result.GetVenue()), nil
}
func (d *Directory) CreateVenue(ctx context.Context, input catalog.CreateVenueInput) (model.Venue, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Venue{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenueResponse, error) {
		return d.client.CreateVenue(ctx, &identityV1.TenantCreateVenueRequest{AccessToken: access, OrganizationId: input.OrganizationID, Name: input.Name, Slug: input.Slug, Address: input.Address, Timezone: input.Timezone})
	}, catalog.ErrVenueSlugConflict)
	if err != nil {
		return model.Venue{}, err
	}
	return venue(result.GetVenue()), nil
}
func (d *Directory) UpdateVenue(ctx context.Context, id string, input catalog.UpdateVenueInput) (model.Venue, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Venue{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenueResponse, error) {
		return d.client.UpdateVenue(ctx, &identityV1.TenantUpdateVenueRequest{AccessToken: access, VenueId: id, Name: input.Name, Slug: input.Slug, Address: input.Address, Timezone: input.Timezone})
	}, catalog.ErrVenueSlugConflict)
	if err != nil {
		return model.Venue{}, err
	}
	return venue(result.GetVenue()), nil
}
func (d *Directory) ArchiveVenue(ctx context.Context, id string) (model.Venue, error) {
	access, err := token(ctx)
	if err != nil {
		return model.Venue{}, err
	}
	result, err := call(ctx, func(ctx context.Context) (*identityV1.TenantVenueResponse, error) {
		return d.client.ArchiveVenue(ctx, &identityV1.TenantArchiveVenueRequest{AccessToken: access, VenueId: id})
	}, nil)
	if err != nil {
		return model.Venue{}, err
	}
	return venue(result.GetVenue()), nil
}
