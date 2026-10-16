package apperror

import (
	"net/http"
	"testing"

	"github.com/kainguyen/badminton-hub/be/services/identity/domain/auth"
)

func TestResolveAuthErrors(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{err: auth.ErrInvalidInput, wantStatus: http.StatusBadRequest, wantCode: "INVALID_REQUEST"},
		{err: auth.ErrEmailTaken, wantStatus: http.StatusConflict, wantCode: "EMAIL_ALREADY_REGISTERED"},
		{err: auth.ErrInvalidCredentials, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_CREDENTIALS"},
		{err: auth.ErrInvalidToken, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_TOKEN"},
		{err: auth.ErrUserInactive, wantStatus: http.StatusForbidden, wantCode: "USER_INACTIVE"},
		{err: auth.ErrUserNotFound, wantStatus: http.StatusNotFound, wantCode: "USER_NOT_FOUND"},
	}
	for _, test := range tests {
		resolved := Resolve(test.err)
		if resolved.Status != test.wantStatus || resolved.Code != test.wantCode {
			t.Fatalf("Resolve(%v) = status %d code %q, want status %d code %q", test.err, resolved.Status, resolved.Code, test.wantStatus, test.wantCode)
		}
	}
}
