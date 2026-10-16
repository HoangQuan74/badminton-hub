package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/services/identity/transport/http/apperror"
	localization "github.com/kainguyen/badminton-hub/be/services/identity/transport/http/i18n"
	"github.com/kainguyen/badminton-hub/be/services/identity/transport/http/models/response"
)

type Endpoint func(*gin.Context) error
type JSONEndpoint[T any] func(*gin.Context, T) error

func Handle(logger *slog.Logger, endpoint Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := endpoint(c); err != nil {
			writeEndpointError(c, logger, err)
		}
	}
}

func HandleJSON[T any](logger *slog.Logger, endpoint JSONEndpoint[T]) gin.HandlerFunc {
	return Handle(logger, func(c *gin.Context) error {
		var input T
		if err := c.ShouldBindJSON(&input); err != nil {
			return apperror.InvalidRequest(err)
		}
		return endpoint(c, input)
	})
}

func writeEndpointError(c *gin.Context, logger *slog.Logger, err error) {
	resolved := apperror.Resolve(err)
	if resolved.Status >= http.StatusInternalServerError {
		logger.Error("HTTP endpoint failed", "error", err, "path", c.FullPath(), "status", resolved.Status)
	}
	c.AbortWithStatusJSON(resolved.Status, response.ErrorEnvelope{
		Error: response.ErrorDetail{Code: resolved.Code, Message: localization.T(c, resolved.MessageID)},
	})
}
