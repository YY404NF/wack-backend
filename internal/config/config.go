package config

import "os"

type Config struct {
	Host            string
	Port            string
	DatabasePath    string
	JWTSecret       string
	CORSAllowOrigin string
}

func Load() Config {
	return Config{
		Host:            getEnv("WACK_HOST", ""),
		Port:            getEnv("WACK_PORT", "8080"),
		DatabasePath:    getEnv("WACK_DB_PATH", "data/wack.db"),
		JWTSecret:       getEnv("WACK_JWT_SECRET", "wack-dev-secret"),
		CORSAllowOrigin: getEnv("WACK_CORS_ALLOW_ORIGIN", "http://8.159.159.150"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
