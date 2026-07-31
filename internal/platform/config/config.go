package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Environment  string
	HTTPAddr     string
	PostgresDSN  string
	RedisAddr    string
	AutoMigrate  bool
	BuildVersion string
}

func Load() (Config, error) {
	cfg := Config{
		Environment:  value("TRAVELINGHUB_ENV", "development"),
		HTTPAddr:     value("TRAVELINGHUB_HTTP_ADDR", ":8080"),
		PostgresDSN:  strings.TrimSpace(os.Getenv("TRAVELINGHUB_POSTGRES_DSN")),
		RedisAddr:    strings.TrimSpace(os.Getenv("TRAVELINGHUB_REDIS_ADDR")),
		AutoMigrate:  value("TRAVELINGHUB_AUTO_MIGRATE", "false") == "true",
		BuildVersion: value("TRAVELINGHUB_BUILD_VERSION", "dev"),
	}
	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("TRAVELINGHUB_POSTGRES_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("TRAVELINGHUB_REDIS_ADDR is required")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if got := strings.TrimSpace(os.Getenv(key)); got != "" {
		return got
	}
	return fallback
}
