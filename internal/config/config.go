package config

import "os"

type Config struct {
	Port         string
	DataDir      string
	DatabasePath string
	JWTSecret    string
}

func Load() Config {
	dataDir := getEnv("WACK_DATA_DIR", "data")
	return Config{
		Port:         getEnv("WACK_PORT", "8080"),
		DataDir:      dataDir,
		DatabasePath: getEnv("WACK_DB_PATH", dataDir+"/wack.db"),
		JWTSecret:    getEnv("WACK_JWT_SECRET", "wack-dev-secret"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
