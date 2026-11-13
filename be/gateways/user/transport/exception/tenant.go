package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TenantFieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type tenantAPIError struct {
	Code      string             `json:"code"`
	Message   string             `json:"message"`
	RequestID string             `json:"request_id"`
	Fields    []TenantFieldError `json:"fields,omitempty"`
}

func WriteTenantError(c *gin.Context, httpStatus int, code, lang string, fields []TenantFieldError) {
	for i := range fields {
		fields[i].Message = TenantMessage(fields[i].Code, lang)
	}
	c.JSON(httpStatus, struct {
		Error tenantAPIError `json:"error"`
	}{Error: tenantAPIError{
		Code: code, Message: TenantMessage(code, lang), RequestID: c.Writer.Header().Get("X-Request-ID"), Fields: fields,
	}})
}

func WriteTenantGRPCError(c *gin.Context, err error, lang string) {
	// Status and allowlisted reasons determine codes; upstream text is never returned.
	value := status.Convert(err)
	httpStatus, code := http.StatusInternalServerError, "INTERNAL_ERROR"
	switch value.Code() {
	case codes.InvalidArgument:
		httpStatus, code = http.StatusBadRequest, "INVALID_REQUEST"
	case codes.Unauthenticated:
		httpStatus, code = http.StatusUnauthorized, "UNAUTHORIZED"
	case codes.PermissionDenied:
		httpStatus, code = http.StatusForbidden, "FORBIDDEN"
	case codes.Unavailable:
		httpStatus, code = http.StatusBadGateway, "DEPENDENCY_UNAVAILABLE"
	case codes.DeadlineExceeded:
		httpStatus, code = http.StatusGatewayTimeout, "DEPENDENCY_TIMEOUT"
	}
	fields := []TenantFieldError{}
	for _, detail := range value.Details() {
		switch detail := detail.(type) {
		case *errdetails.ErrorInfo:
			if detail.Domain != "badminton.identity.tenant" {
				continue
			}
			if value.Code() == codes.AlreadyExists && (detail.Reason == "ORGANIZATION_SLUG_ALREADY_EXISTS" || detail.Reason == "IDEMPOTENCY_KEY_CONFLICT") {
				httpStatus, code = http.StatusConflict, detail.Reason
			}
			if value.Code() == codes.FailedPrecondition && detail.Reason == "TENANT_ROLE_CONFLICT" {
				httpStatus, code = http.StatusConflict, detail.Reason
			}
		case *errdetails.BadRequest:
			if value.Code() != codes.InvalidArgument {
				continue
			}
			for _, field := range detail.FieldViolations {
				allowedField := field.Field == "name" || field.Field == "slug" || field.Field == "idempotency_key"
				allowedCode := field.Description == "REQUIRED" || field.Description == "TOO_LONG" || field.Description == "UUID_INVALID" || field.Description == "SLUG_INVALID" || field.Description == "NAME_INVALID"
				if allowedField && allowedCode {
					fields = append(fields, TenantFieldError{Field: field.Field, Code: field.Description})
				}
			}
		}
	}
	if code == "ORGANIZATION_SLUG_ALREADY_EXISTS" {
		fields = append(fields, TenantFieldError{Field: "slug", Code: "SLUG_TAKEN"})
	}
	WriteTenantError(c, httpStatus, code, lang, fields)
}

var OrganizationErrors = ErrorConfig{NotFoundCode: "ORGANIZATION_NOT_FOUND", AlreadyExistsCode: "ORGANIZATION_SLUG_ALREADY_EXISTS", FailedPreconditionCode: "ORGANIZATION_STATE_CONFLICT", UnavailableCode: "DEPENDENCY_UNAVAILABLE"}
