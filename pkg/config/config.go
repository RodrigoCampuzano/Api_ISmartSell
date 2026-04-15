package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DSN             string
	JWTSecret       string
	JWTTTLHrs       int
	MPAccessToken   string
	MPClientID      string
	MPClientSecret  string
	MPWebhookSecret string
	MPRedirectURI   string
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		DSN:             getEnv("DSN", "postgres://postgres:postgres@localhost:5432/pos_app?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		JWTTTLHrs:       getEnvInt("JWT_TTL_HOURS", 72),
		MPAccessToken:   getEnv("MP_ACCESS_TOKEN", ""),
		MPClientID:      getEnv("MP_CLIENT_ID", ""),
		MPClientSecret:  getEnv("MP_CLIENT_SECRET", ""),
		MPWebhookSecret: getEnv("MP_WEBHOOK_SECRET", ""),
		MPRedirectURI:   getEnv("MP_REDIRECT_URI", "http://localhost:8080/api/v1/payments/oauth/callback"),
	}
}

func (c Config) Addr() string { return fmt.Sprintf(":%s", c.Port) }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}
