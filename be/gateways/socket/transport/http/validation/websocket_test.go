package validation

import "testing"

func TestAccessToken(t *testing.T) {
	for _, header := range []string{"", "Basic token", "Bearer", "Bearer abc", "Bearer a.b.c.d"} {
		if token, ok := AccessToken(header); ok || token != "" {
			t.Fatalf("accepted malformed header %q", header)
		}
	}
	if token, ok := AccessToken("Bearer header.payload.signature"); !ok || token != "header.payload.signature" {
		t.Fatal("valid token shape rejected")
	}
}
