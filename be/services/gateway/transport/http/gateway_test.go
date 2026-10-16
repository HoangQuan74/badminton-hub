package gatewayhttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gatewayconfig "github.com/kainguyen/badminton-hub/be/services/gateway/config"
)

func TestGatewayProxiesIdentityRoutes(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/auth/login" {
			t.Fatalf("identity received path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("identity received Authorization %q", got)
		}
		if got := r.Header.Get("X-Request-ID"); got == "" {
			t.Fatal("identity did not receive X-Request-ID")
		}
		return jsonResponse(http.StatusOK, `{"proxied":true}`), nil
	})

	engine, err := New(Dependencies{Config: gatewayconfig.Config{IdentityURL: "http://identity:8080", DependencyTimeout: time.Second}, Transport: transport})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com"}`))
	request.Header.Set("Authorization", "Bearer access-token")
	response := newCloseNotifyRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != `{"proxied":true}` {
		t.Fatalf("body = %q", body)
	}
}

func TestGatewayReadinessFollowsIdentity(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/ready" {
			t.Fatalf("identity received path %q", r.URL.Path)
		}
		return jsonResponse(http.StatusOK, `{"status":"ready"}`), nil
	})
	client := &http.Client{Transport: transport}

	engine, err := New(Dependencies{Config: gatewayconfig.Config{IdentityURL: "http://identity:8080", DependencyTimeout: time.Second}, HTTPClient: client, Transport: transport})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	response := newCloseNotifyRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != `{"status":"ready"}` {
		t.Fatalf("body = %q", body)
	}
}

func TestGatewayReturnsBadGatewayWhenIdentityIsUnavailable(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	engine, err := New(Dependencies{Config: gatewayconfig.Config{IdentityURL: "http://identity:8080", DependencyTimeout: 50 * time.Millisecond}, Transport: transport})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	response := newCloseNotifyRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadGateway)
	}
	if body := response.Body.String(); !strings.Contains(body, `"IDENTITY_UNAVAILABLE"`) {
		t.Fatalf("body = %q", body)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
	notify chan bool
}

func newCloseNotifyRecorder() *closeNotifyRecorder {
	return &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder(), notify: make(chan bool)}
}

func (r *closeNotifyRecorder) CloseNotify() <-chan bool {
	return r.notify
}
