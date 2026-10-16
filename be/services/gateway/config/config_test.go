package gatewayconfig

import "testing"

func TestLoadUsesGatewayScopedConfiguration(t *testing.T) {
	t.Setenv("GATEWAY_IDENTITY_URL", "http://identity:8080")
	t.Setenv("GATEWAY_HTTP_ADDR", ":18080")
	t.Setenv("GATEWAY_SHUTDOWN_TIMEOUT", "11s")
	t.Setenv("GATEWAY_DEPENDENCY_TIMEOUT", "2s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != ":18080" || cfg.IdentityURL != "http://identity:8080" {
		t.Fatalf("unexpected gateway config: %+v", cfg)
	}
}
