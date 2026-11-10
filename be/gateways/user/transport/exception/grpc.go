package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorConfig struct {
	NotFoundCode           string
	AlreadyExistsCode      string
	UnavailableCode        string
	FailedPreconditionCode string
}

func WriteGRPC(c *gin.Context, value any, err error, successStatus int, configs ...ErrorConfig) {
	config := ErrorConfig{NotFoundCode: "USER_NOT_FOUND", AlreadyExistsCode: "EMAIL_ALREADY_REGISTERED"}
	if len(configs) > 0 {
		if configs[0].NotFoundCode != "" {
			config.NotFoundCode = configs[0].NotFoundCode
		}
		if configs[0].FailedPreconditionCode != "" {
			config.FailedPreconditionCode = configs[0].FailedPreconditionCode
		}
		if configs[0].UnavailableCode != "" {
			config.UnavailableCode = configs[0].UnavailableCode
		}
		if configs[0].AlreadyExistsCode != "" {
			config.AlreadyExistsCode = configs[0].AlreadyExistsCode
		}
	}
	if err == nil {
		if successStatus == http.StatusNoContent {
			c.Status(successStatus)
			return
		}
		c.JSON(successStatus, value)
		return
	}
	code := status.Code(err)
	httpStatus, apiCode := http.StatusInternalServerError, "INTERNAL_ERROR"
	switch code {
	case codes.InvalidArgument:
		httpStatus, apiCode = http.StatusBadRequest, "INVALID_REQUEST"
	case codes.AlreadyExists:
		httpStatus, apiCode = http.StatusConflict, config.AlreadyExistsCode
	case codes.FailedPrecondition:
		httpStatus, apiCode = http.StatusConflict, "INVALID_STATE"
		if config.FailedPreconditionCode != "" {
			apiCode = config.FailedPreconditionCode
		}
	case codes.Unauthenticated:
		httpStatus, apiCode = http.StatusUnauthorized, "UNAUTHORIZED"
	case codes.PermissionDenied:
		httpStatus, apiCode = http.StatusForbidden, "FORBIDDEN"
	case codes.NotFound:
		httpStatus, apiCode = http.StatusNotFound, config.NotFoundCode
	case codes.Unavailable:
		httpStatus, apiCode = http.StatusBadGateway, "IDENTITY_UNAVAILABLE"
		if config.UnavailableCode != "" {
			apiCode = config.UnavailableCode
		}
	}
	c.JSON(httpStatus, gin.H{"error": gin.H{"code": apiCode, "message": err.Error()}})
}
