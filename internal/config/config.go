package config

import (
	"os"
	"strconv"
)

type Config struct {
	Host            string
	Port            string
	DatabasePath    string
	JWTSecret       string
	CORSAllowOrigin string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
}

func Load() Config {
	return Config{
		Host:            getEnv("WACK_HOST", ""),
		Port:            getEnv("WACK_PORT", "8080"),
		DatabasePath:    getEnv("WACK_DB_PATH", "data/wack.db"),
		JWTSecret:       getEnv("WACK_JWT_SECRET", "wack-dev-secret"),
		CORSAllowOrigin: getEnv("WACK_CORS_ALLOW_ORIGIN", "http://8.159.159.150"),
		RedisAddr:       getEnv("WACK_REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:   getEnv("WACK_REDIS_PASSWORD", ""),
		RedisDB:         getEnvInt("WACK_REDIS_DB", 0),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
