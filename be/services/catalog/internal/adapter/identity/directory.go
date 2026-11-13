package identityadapter

import (
	"context"
	"strings"

	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/constant"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogException "github.com/kainguyen/badminton-hub/be/services/catalog/internal/exception"
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/model"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
	"google.golang.org/grpc/metadata"
)

type Directory struct {
	client identityV1.IdentityVenueServiceClient
}

func NewDirectory(client identityV1.IdentityVenueServiceClient) *Directory {
	return &Directory{client: client}
}

func parseAccessToken(ctx context.Context) (string, error) {
	incoming, _ := metadata.FromIncomingContext(ctx)
	values := incoming.Get("authorization")
	if len(values) != 1 {
		return "", catalogException.ErrUnauthenticated
	}
	scheme, value, ok := strings.Cut(values[0], " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(value) == "" {
		return "", catalogException.ErrUnauthenticated
	}

	return strings.TrimSpace(value), nil
}

func (d *Directory) ListPublicVenues(ctx context.Context) ([]model.Venue, error) {
	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	result, err := d.client.ListPublicVenues(rpcContext, &identityV1.TenantListPublicVenuesRequest{})
	if err != nil {
		return nil, catalogException.IdentityError(err, nil)
	}

	return mapVenues(result.GetItems()), nil
}

func (d *Directory) GetPublicVenue(ctx context.Context, slug string) (model.Venue, error) {
	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	result, err := d.client.GetPublicVenue(rpcContext, &identityV1.TenantGetPublicVenueRequest{Slug: slug})
	if err != nil {
		return model.Venue{}, catalogException.IdentityError(err, nil)
	}

	return mapVenue(result.GetVenue()), nil
}

func (d *Directory) GetPublicVenueByID(ctx context.Context, id string) (model.Venue, error) {
	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	result, err := d.client.GetPublicVenueByID(rpcContext, &identityV1.TenantGetPublicVenueByIDRequest{VenueId: id})
	if err != nil {
		return model.Venue{}, catalogException.IdentityError(err, nil)
	}

	return mapVenue(result.GetVenue()), nil
}

func (d *Directory) ListVenues(ctx context.Context) ([]model.Venue, error) {
	access, err := parseAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	result, err := d.client.ListVenues(rpcContext, &identityV1.TenantListVenuesRequest{AccessToken: access})
	if err != nil {
		return nil, catalogException.IdentityError(err, nil)
	}

	return mapVenues(result.GetItems()), nil
}

func (d *Directory) AuthorizeVenue(ctx context.Context, id string, requireOwner bool) (model.Venue, error) {
	access, err := parseAccessToken(ctx)
	if err != nil {
		return model.Venue{}, err
	}

	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	request := &identityV1.TenantAuthorizeVenueRequest{
		AccessToken:  access,
		VenueId:      id,
		RequireOwner: requireOwner,
	}

	result, err := d.client.AuthorizeVenue(rpcContext, request)
	if err != nil {
		return model.Venue{}, catalogException.IdentityError(err, nil)
	}

	return mapVenue(result.GetVenue()), nil
}

func (d *Directory) CreateVenue(ctx context.Context, input catalog.CreateVenueInput) (model.Venue, error) {
	access, err := parseAccessToken(ctx)
	if err != nil {
		return model.Venue{}, err
	}

	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	request := &identityV1.TenantCreateVenueRequest{
		AccessToken:    access,
		OrganizationId: input.OrganizationID,
		Name:           input.Name,
		Slug:           input.Slug,
		Address:        input.Address,
		Timezone:       input.Timezone,
	}

	result, err := d.client.CreateVenue(rpcContext, request)
	if err != nil {
		return model.Venue{}, catalogException.IdentityError(err, catalogException.ErrVenueSlugConflict)
	}

	return mapVenue(result.GetVenue()), nil
}

func (d *Directory) UpdateVenue(ctx context.Context, id string, input catalog.UpdateVenueInput) (model.Venue, error) {
	access, err := parseAccessToken(ctx)
	if err != nil {
		return model.Venue{}, err
	}

	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	request := &identityV1.TenantUpdateVenueRequest{
		AccessToken: access,
		VenueId:     id,
		Name:        input.Name,
		Slug:        input.Slug,
		Address:     input.Address,
		Timezone:    input.Timezone,
	}

	result, err := d.client.UpdateVenue(rpcContext, request)
	if err != nil {
		return model.Venue{}, catalogException.IdentityError(err, catalogException.ErrVenueSlugConflict)
	}

	return mapVenue(result.GetVenue()), nil
}

func (d *Directory) ArchiveVenue(ctx context.Context, id string) (model.Venue, error) {
	access, err := parseAccessToken(ctx)
	if err != nil {
		return model.Venue{}, err
	}

	rpcContext, cancel := context.WithTimeout(ctx, constant.IdentityRPCTimeout)
	defer cancel()

	request := &identityV1.TenantArchiveVenueRequest{
		AccessToken: access,
		VenueId:     id,
	}

	result, err := d.client.ArchiveVenue(rpcContext, request)
	if err != nil {
		return model.Venue{}, catalogException.IdentityError(err, nil)
	}

	return mapVenue(result.GetVenue()), nil
}
