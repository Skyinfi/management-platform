package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr        string
	Env         string
	LogLevel    string
	EnableCORS  bool
	JWTSecret   string
	AllowOrigin string
}

func Load() Config {
	return Config{
		Addr:        getString("APP_MANAGER_ADDR", ":8080"),
		Env:         getString("APP_MANAGER_ENV", "development"),
		LogLevel:    getString("APP_MANAGER_LOG_LEVEL", "info"),
		EnableCORS:  getBool("APP_MANAGER_ENABLE_CORS", true),
		JWTSecret:   getString("APP_MANAGER_JWT_SECRET", "dev-secret"),
		AllowOrigin: getString("APP_MANAGER_ALLOW_ORIGIN", "*"),
	}
}

func getString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
