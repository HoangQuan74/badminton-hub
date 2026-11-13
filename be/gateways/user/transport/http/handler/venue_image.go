package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/gateways/user/transport/exception"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	catalogV1 "github.com/kainguyen/badminton-hub/contracts/gen/catalog/v1"
)

const maxImageBytes = 2 * 1024 * 1024
const maxMultipartBytes = maxImageBytes + 64*1024

func (h *CatalogHandler) UploadVenueImage(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusCreated,
			exception.ImageErrors,
		)
		return
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxMultipartBytes,
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		exception.WriteImageRequired(c)
		return
	}

	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxImageBytes+1))
	if err != nil || len(content) == 0 || len(content) > maxImageBytes {
		exception.WriteImageInvalid(c)
		return
	}

	result, err := h.catalog.UploadVenueImage(catalogTenantContext(c), &catalogV1.UploadVenueImageRequest{
		VenueId: c.Param("venue_id"),
		Content: content,
	})
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusCreated,
			exception.ImageErrors,
		)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"image": imageResponse(result.GetImage(), true)})
}

func (h *CatalogHandler) ListPublicVenueImages(c *gin.Context) {
	result, err := h.catalog.ListPublicVenueImages(request.RPCContext(c, platformService.UserGateway), &catalogV1.ListPublicVenueImagesRequest{VenueId: c.Param("venue_id")})
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
			exception.ImageErrors,
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": imageResponses(result.GetItems(), false)})
}

func (h *CatalogHandler) ListTenantVenueImages(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
			exception.ImageErrors,
		)
		return
	}

	result, err := h.catalog.ListTenantVenueImages(catalogTenantContext(c), &catalogV1.ListTenantVenueImagesRequest{VenueId: c.Param("venue_id")})
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
			exception.ImageErrors,
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": imageResponses(result.GetItems(), true)})
}

func (h *CatalogHandler) GetPublicVenueImage(c *gin.Context) {
	result, err := h.catalog.GetPublicVenueImage(request.RPCContext(c, platformService.UserGateway), &catalogV1.GetVenueImageRequest{ImageId: c.Param("image_id")})
	writeImage(
		c,
		result,
		err,
	)
}

func (h *CatalogHandler) GetTenantVenueImage(c *gin.Context) {
	if _, err := h.tenantActorID(c); err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
			exception.ImageErrors,
		)
		return
	}

	result, err := h.catalog.GetTenantVenueImage(catalogTenantContext(c), &catalogV1.GetVenueImageRequest{ImageId: c.Param("image_id")})
	writeImage(
		c,
		result,
		err,
	)
}

func writeImage(c *gin.Context, result *catalogV1.VenueImageContentResponse, err error) {
	if err != nil {
		exception.WriteGRPC(
			c,
			nil,
			err,
			http.StatusOK,
			exception.ImageErrors,
		)
		return
	}

	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-store")
	c.Data(
		http.StatusOK,
		result.GetContentType(),
		result.GetContent(),
	)
}

func imageResponses(items []*catalogV1.VenueImage, tenant bool) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, imageResponse(item, tenant))
	}

	return result
}

func imageResponse(item *catalogV1.VenueImage, tenant bool) gin.H {
	prefix := "/api/v1/venue-images/"
	if tenant {
		prefix = "/api/v1/tenant/venue-images/"
	}

	return gin.H{
		"id":           item.GetId(),
		"venue_id":     item.GetVenueId(),
		"content_type": item.GetContentType(),
		"size_bytes":   item.GetSizeBytes(),
		"created_at":   item.GetCreatedAt(),
		"url":          prefix + item.GetId(),
	}
}
