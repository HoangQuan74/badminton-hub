// Package request contains reusable HTTP input and request-context helpers.
package request

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	grpcMetadata "github.com/kainguyen/badminton-hub/be/platform/grpc/metadata"
	httpContext "github.com/kainguyen/badminton-hub/be/platform/http/context"
)

func BindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid request."}})
		return true
	}
	return false
}

func BearerToken(c *gin.Context) string {
	return strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
}

func Param(c *gin.Context, name string) string {
	return strings.TrimSpace(c.Param(name))
}

func QueryInt32(c *gin.Context, name string, fallback int32) int32 {
	value, err := strconv.ParseInt(c.DefaultQuery(name, strconv.FormatInt(int64(fallback), 10)), 10, 32)
	if err != nil || value < 0 {
		return fallback
	}
	return int32(value)
}

type ListParams struct {
	Limit       int32
	Offset      int32
	Search      string
	CreatedFrom string
	CreatedTo   string
}

func ListQuery(c *gin.Context) (ListParams, error) {
	params := ListParams{
		Limit:       QueryInt32(c, "limit", 20),
		Offset:      QueryInt32(c, "offset", 0),
		Search:      strings.TrimSpace(c.Query("search")),
		CreatedFrom: strings.TrimSpace(c.Query("created_from")),
		CreatedTo:   strings.TrimSpace(c.Query("created_to")),
	}
	if err := validateDate(params.CreatedFrom); err != nil {
		return ListParams{}, fmt.Errorf("created_from: %w", err)
	}
	if err := validateDate(params.CreatedTo); err != nil {
		return ListParams{}, fmt.Errorf("created_to: %w", err)
	}
	return params, nil
}

func validateDate(value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("must be an RFC3339 timestamp or YYYY-MM-DD date")
	}
	return nil
}

func RPCContext(c *gin.Context, serviceIdentity string) context.Context {
	metadata, _ := httpContext.RequestMetadataFromContext(c.Request.Context())
	return grpcMetadata.WithRequestMetadata(c.Request.Context(), grpcMetadata.RequestMetadata{
		RequestID:       metadata.RequestID,
		CorrelationID:   metadata.CorrelationID,
		ServiceIdentity: serviceIdentity,
	})
}
