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
	if cfg.HTTPAddr != ":8080" || cfg.Environment != "development" {
		t.Fatalf("Load() = %#v, want default HTTP address and environment", cfg)
	}
}
