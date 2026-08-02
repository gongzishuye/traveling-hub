package config

import "testing"

func TestLoadRejectsMissingPostgresDSN(t *testing.T) {
	t.Setenv("TRAVELINGHUB_POSTGRES_DSN", "")
	t.Setenv("TRAVELINGHUB_REDIS_ADDR", "localhost:6379")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing Postgres DSN error")
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("TRAVELINGHUB_POSTGRES_DSN", "postgres://example")
	t.Setenv("TRAVELINGHUB_REDIS_ADDR", "localhost:6379")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.Environment != "development" || cfg.WebOrigin != "http://127.0.0.1:5173" {
		t.Fatalf("Load() = %#v, want default HTTP address and environment", cfg)
	}
}

func TestLoadRequiresHTTPSWebOriginInProduction(t *testing.T) {
	t.Setenv("TRAVELINGHUB_POSTGRES_DSN", "postgres://example")
	t.Setenv("TRAVELINGHUB_REDIS_ADDR", "localhost:6379")
	t.Setenv("TRAVELINGHUB_ENV", "production")
	t.Setenv("TRAVELINGHUB_WEB_ORIGIN", "http://example.com")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production HTTP origin rejection")
	}
}

func TestLoadRejectsProductionEmailVerificationBypass(t *testing.T) {
	t.Setenv("TRAVELINGHUB_POSTGRES_DSN", "postgres://example")
	t.Setenv("TRAVELINGHUB_REDIS_ADDR", "localhost:6379")
	t.Setenv("TRAVELINGHUB_ENV", "production")
	t.Setenv("TRAVELINGHUB_WEB_ORIGIN", "https://example.com")
	t.Setenv("TRAVELINGHUB_AUTO_VERIFY_EMAIL", "true")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want production email verification bypass rejection")
	}
}

func TestLoadRequiresSMTPConfigurationInProduction(t *testing.T) {
	t.Setenv("TRAVELINGHUB_POSTGRES_DSN", "postgres://example")
	t.Setenv("TRAVELINGHUB_REDIS_ADDR", "localhost:6379")
	t.Setenv("TRAVELINGHUB_ENV", "production")
	t.Setenv("TRAVELINGHUB_WEB_ORIGIN", "https://example.com")
	t.Setenv("TRAVELINGHUB_AUTO_VERIFY_EMAIL", "false")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing SMTP configuration rejection")
	}
}
