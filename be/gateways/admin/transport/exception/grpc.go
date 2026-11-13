package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorConfig struct{ NotFoundCode, AlreadyExistsCode, FailedPreconditionCode string }

var PermissionErrors = ErrorConfig{NotFoundCode: "PERMISSION_NOT_FOUND", AlreadyExistsCode: "PERMISSION_ALREADY_EXISTS"}

func WriteGRPC(c *gin.Context, value any, err error, successStatus int, configs ...ErrorConfig) {
	config := ErrorConfig{NotFoundCode: "RESOURCE_NOT_FOUND"}
	if len(configs) > 0 {
		config = configs[0]
	}
	if err == nil {
		if successStatus == http.StatusNoContent {
			c.Status(successStatus)
			return
		}
		c.JSON(successStatus, value)
		return
	}
	grpcCode := status.Code(err)
	httpStatus, apiCode := http.StatusInternalServerError, "INTERNAL_ERROR"
	switch grpcCode {
	case codes.InvalidArgument:
		httpStatus, apiCode = http.StatusBadRequest, "INVALID_REQUEST"
	case codes.AlreadyExists:
		httpStatus, apiCode = http.StatusConflict, "EMAIL_ALREADY_REGISTERED"
		if config.AlreadyExistsCode != "" {
			apiCode = config.AlreadyExistsCode
		}
	case codes.Unauthenticated:
		httpStatus, apiCode = http.StatusUnauthorized, "UNAUTHORIZED"
	case codes.PermissionDenied:
		httpStatus, apiCode = http.StatusForbidden, "FORBIDDEN"
	case codes.NotFound:
		httpStatus, apiCode = http.StatusNotFound, config.NotFoundCode
	case codes.FailedPrecondition:
		httpStatus, apiCode = http.StatusConflict, "STATE_CONFLICT"
		if config.FailedPreconditionCode != "" {
			apiCode = config.FailedPreconditionCode
		}
	case codes.Unimplemented:
		httpStatus, apiCode = http.StatusNotImplemented, "FEATURE_NOT_IMPLEMENTED"
	case codes.Unavailable:
		httpStatus, apiCode = http.StatusBadGateway, "IDENTITY_UNAVAILABLE"
	}
	c.JSON(httpStatus, gin.H{"error": gin.H{"code": apiCode, "message": err.Error()}})
}
