package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port       string
	DSN        string
	JWTSecret  string
	JWTTTLHrs  int
}

func Load() Config {
	return Config{
		Port:      getEnv("PORT", "8080"),
		DSN:       getEnv("DSN", "root:root@tcp(localhost:3306)/pos_app?parseTime=true&charset=utf8mb4"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		JWTTTLHrs: getEnvInt("JWT_TTL_HOURS", 72),
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
