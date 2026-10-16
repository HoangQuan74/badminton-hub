package identityconfig

import "testing"

func TestLoadUsesIdentityScopedConfiguration(t *testing.T) {
	t.Setenv("IDENTITY_DATABASE_URL", "postgres://identity")
	t.Setenv("IDENTITY_REDIS_URL", "redis://identity-redis:6379/1")
	t.Setenv("IDENTITY_JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("IDENTITY_HTTP_ADDR", ":18081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabaseURL != "postgres://identity" || cfg.RedisURL != "redis://identity-redis:6379/1" {
		t.Fatalf("unexpected identity storage config: %+v", cfg)
	}
}
