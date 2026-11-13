package cataloggrpc

import (
	"github.com/kainguyen/badminton-hub/be/services/catalog/internal/domain/catalog"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

type CatalogServer struct {
	catalogV1.UnimplementedCatalogServiceServer
	service  *catalog.Service
	identity identityV1.IdentityTenantAuthServiceClient
	images   *catalog.ImageService
}

func NewCatalogServer(service *catalog.Service, identity identityV1.IdentityTenantAuthServiceClient, imageServices ...*catalog.ImageService) *CatalogServer {
	server := &CatalogServer{service: service, identity: identity}
	if len(imageServices) > 0 {
		server.images = imageServices[0]
	}
	return server
}
