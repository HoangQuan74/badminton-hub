package middleware

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/services/identity/domain/auth"
	localization "github.com/kainguyen/badminton-hub/be/services/identity/transport/http/i18n"
)

type endpointRequest struct {
	Name string `json:"name" binding:"required"`
}

func TestHandleJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	translator, err := localization.New()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		body       string
		language   string
		endpoint   JSONEndpoint[endpointRequest]
		wantStatus int
		wantBody   string
	}{
		{name: "valid", body: `{"name":"Court Owner"}`, endpoint: func(c *gin.Context, input endpointRequest) error { c.Status(http.StatusNoContent); return nil }, wantStatus: http.StatusNoContent},
		{name: "invalid JSON", body: `{}`, language: "vi", endpoint: func(*gin.Context, endpointRequest) error { return nil }, wantStatus: http.StatusBadRequest, wantBody: "Dữ liệu yêu cầu không hợp lệ."},
		{name: "domain error", body: `{"name":"Court Owner"}`, endpoint: func(*gin.Context, endpointRequest) error { return auth.ErrEmailTaken }, wantStatus: http.StatusConflict, wantBody: "EMAIL_ALREADY_REGISTERED"},
		{name: "wrapped domain error", body: `{"name":"Court Owner"}`, endpoint: func(*gin.Context, endpointRequest) error {
			return errors.Join(errors.New("repository"), auth.ErrInvalidCredentials)
		}, wantStatus: http.StatusUnauthorized, wantBody: "INVALID_CREDENTIALS"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(translator.Middleware())
			engine.POST("/test", HandleJSON(logger, test.endpoint))
			request := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Accept-Language", test.language)
			writer := httptest.NewRecorder()
			engine.ServeHTTP(writer, request)
			if writer.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", writer.Code, test.wantStatus)
			}
			if test.wantBody != "" && !strings.Contains(writer.Body.String(), test.wantBody) {
				t.Fatalf("body = %q, want it to contain %q", writer.Body.String(), test.wantBody)
			}
		})
	}
}
