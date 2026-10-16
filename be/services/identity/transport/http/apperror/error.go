package apperror

import (
	"errors"
	"net/http"

	"github.com/kainguyen/badminton-hub/be/services/identity/domain/auth"
)

type Error struct {
	Status    int
	Code      string
	MessageID string
	Cause     error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return e.MessageID
}

func (e *Error) Unwrap() error { return e.Cause }

func InvalidRequest(cause error) *Error {
	return &Error{Status: http.StatusBadRequest, Code: "INVALID_REQUEST", MessageID: "error.invalid_request", Cause: cause}
}

func Unauthorized() *Error {
	return &Error{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", MessageID: "error.unauthorized"}
}

func Resolve(err error) *Error {
	var apiError *Error
	if errors.As(err, &apiError) {
		return apiError
	}
	switch {
	case errors.Is(err, auth.ErrInvalidInput):
		return &Error{Status: http.StatusBadRequest, Code: "INVALID_REQUEST", MessageID: auth.ErrInvalidInput.Code, Cause: err}
	case errors.Is(err, auth.ErrEmailTaken):
		return &Error{Status: http.StatusConflict, Code: "EMAIL_ALREADY_REGISTERED", MessageID: auth.ErrEmailTaken.Code, Cause: err}
	case errors.Is(err, auth.ErrInvalidCredentials):
		return &Error{Status: http.StatusUnauthorized, Code: "INVALID_CREDENTIALS", MessageID: auth.ErrInvalidCredentials.Code, Cause: err}
	case errors.Is(err, auth.ErrInvalidToken):
		return &Error{Status: http.StatusUnauthorized, Code: "INVALID_TOKEN", MessageID: auth.ErrInvalidToken.Code, Cause: err}
	case errors.Is(err, auth.ErrUserInactive):
		return &Error{Status: http.StatusForbidden, Code: "USER_INACTIVE", MessageID: auth.ErrUserInactive.Code, Cause: err}
	case errors.Is(err, auth.ErrUserNotFound):
		return &Error{Status: http.StatusNotFound, Code: "USER_NOT_FOUND", MessageID: auth.ErrUserNotFound.Code, Cause: err}
	default:
		return &Error{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", MessageID: "error.internal", Cause: err}
	}
}
