package config

import (
	"os"
	"time"
)

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

type Config struct {
	JWTSecret     string
	DBPath        string
	Port          string
	AdminEmail    string
	AdminPassword string
	AdminFullname string
	GoRunnerImage string
	RunTimeout    time.Duration
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "playground.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	runTimeout := 60 * time.Second
	if v := os.Getenv("RUN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			runTimeout = d
		}
	}

	return &Config{
		JWTSecret:     jwtSecret,
		DBPath:        dbPath,
		Port:          port,
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		AdminFullname: getEnvOrDefault("ADMIN_FULLNAME", "Admin"),
		GoRunnerImage: getEnvOrDefault("GO_RUNNER_IMAGE", "golang:1.21-alpine"),
		RunTimeout:    runTimeout,
	}
}
