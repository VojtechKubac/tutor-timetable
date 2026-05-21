package config

import "os"

type Config struct {
	DatabaseURL  string
	JWTSecret    string
	FrontendURL  string
	Port         string
	SeedEmail    string
	SeedPassword string
	SeedName     string
}

func Load() *Config {
	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://timetable:secret@localhost:5432/timetable?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:5173"),
		Port:         getEnv("PORT", "8080"),
		SeedEmail:    getEnv("SEED_EMAIL", "teacher@example.com"),
		SeedPassword: getEnv("SEED_PASSWORD", "changeme"),
		SeedName:     getEnv("SEED_NAME", "Music Teacher"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
