package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppMode      string
	HTTPPort     string
	DatabaseURL  string
	RedisAddress string
	JWTSecret    string
	JWTExpiry    time.Duration
}

func Load() Config {
	return Config{
		AppMode:      getEnv("APP_MODE", "api"),
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=event_booking port=5432 sslmode=disable"),
		RedisAddress: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-secret"),
		JWTExpiry:    getEnvDuration("JWT_EXPIRY_HOURS", 24) * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvDuration(key string, fallback int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallback)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return time.Duration(fallback)
	}

	return time.Duration(parsed)
}
