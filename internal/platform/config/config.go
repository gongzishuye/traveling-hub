package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment           string
	HTTPAddr              string
	PostgresDSN           string
	RedisAddr             string
	AutoMigrate           bool
	BuildVersion          string
	SessionTTL            time.Duration
	AutoVerifyEmail       bool
	WebOrigin             string
	SMTPAddr              string
	SMTPFrom              string
	SMTPUsername          string
	SMTPPassword          string
	JourneyWorkerInterval time.Duration
	JourneyWorkerBatch    int
}

func Load() (Config, error) {
	cfg := Config{
		Environment:           value("TRAVELINGHUB_ENV", "development"),
		HTTPAddr:              value("TRAVELINGHUB_HTTP_ADDR", ":8080"),
		PostgresDSN:           strings.TrimSpace(os.Getenv("TRAVELINGHUB_POSTGRES_DSN")),
		RedisAddr:             strings.TrimSpace(os.Getenv("TRAVELINGHUB_REDIS_ADDR")),
		AutoMigrate:           value("TRAVELINGHUB_AUTO_MIGRATE", "false") == "true",
		BuildVersion:          value("TRAVELINGHUB_BUILD_VERSION", "dev"),
		SessionTTL:            duration("TRAVELINGHUB_SESSION_TTL", 7*24*time.Hour),
		AutoVerifyEmail:       value("TRAVELINGHUB_AUTO_VERIFY_EMAIL", "false") == "true",
		WebOrigin:             value("TRAVELINGHUB_WEB_ORIGIN", "http://localhost:5173"),
		SMTPAddr:              strings.TrimSpace(os.Getenv("TRAVELINGHUB_SMTP_ADDR")),
		SMTPFrom:              strings.TrimSpace(os.Getenv("TRAVELINGHUB_SMTP_FROM")),
		SMTPUsername:          strings.TrimSpace(os.Getenv("TRAVELINGHUB_SMTP_USERNAME")),
		SMTPPassword:          strings.TrimSpace(os.Getenv("TRAVELINGHUB_SMTP_PASSWORD")),
		JourneyWorkerInterval: duration("TRAVELINGHUB_JOURNEY_WORKER_INTERVAL", time.Minute),
		JourneyWorkerBatch:    positiveInt("TRAVELINGHUB_JOURNEY_WORKER_BATCH", 100),
	}
	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("TRAVELINGHUB_POSTGRES_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("TRAVELINGHUB_REDIS_ADDR is required")
	}
	if cfg.Environment == "production" && !strings.HasPrefix(cfg.WebOrigin, "https://") {
		return Config{}, fmt.Errorf("TRAVELINGHUB_WEB_ORIGIN must use https in production")
	}
	if cfg.Environment == "production" && cfg.AutoVerifyEmail {
		return Config{}, fmt.Errorf("TRAVELINGHUB_AUTO_VERIFY_EMAIL must be false in production")
	}
	if cfg.Environment == "production" && (cfg.SMTPAddr == "" || cfg.SMTPFrom == "") {
		return Config{}, fmt.Errorf("TRAVELINGHUB_SMTP_ADDR and TRAVELINGHUB_SMTP_FROM are required in production")
	}
	return cfg, nil
}

func positiveInt(key string, fallback int) int {
	parsed, err := strconv.Atoi(value(key, strconv.Itoa(fallback)))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func duration(key string, fallback time.Duration) time.Duration {
	raw := value(key, fallback.String())
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func value(key, fallback string) string {
	if got := strings.TrimSpace(os.Getenv(key)); got != "" {
		return got
	}
	return fallback
}
