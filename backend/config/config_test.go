package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv(t,
		"DATABASE_URL", "JWT_SECRET", "FRONTEND_URL", "PORT",
		"SEED_EMAIL", "SEED_PASSWORD", "SEED_NAME",
	)

	cfg := Load()

	if cfg.Port != "8081" {
		t.Fatalf("Port: got %q want 8081", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://timetable:secret@localhost:5432/timetable?sslmode=disable" {
		t.Fatalf("DatabaseURL: got %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "dev-secret-change-in-production" {
		t.Fatalf("JWTSecret: got %q", cfg.JWTSecret)
	}
	if cfg.FrontendURL != "http://localhost:3001" {
		t.Fatalf("FrontendURL: got %q want http://localhost:3001", cfg.FrontendURL)
	}
	if cfg.SeedEmail != "teacher@example.com" {
		t.Fatalf("SeedEmail: got %q", cfg.SeedEmail)
	}
	if cfg.SeedPassword != "changeme" {
		t.Fatalf("SeedPassword: got %q", cfg.SeedPassword)
	}
	if cfg.SeedName != "Music Teacher" {
		t.Fatalf("SeedName: got %q", cfg.SeedName)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("FRONTEND_URL", "http://example.test:3001")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg := Load()

	if cfg.Port != "9999" {
		t.Fatalf("Port: got %q want 9999", cfg.Port)
	}
	if cfg.FrontendURL != "http://example.test:3001" {
		t.Fatalf("FrontendURL: got %q", cfg.FrontendURL)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Fatalf("JWTSecret: got %q", cfg.JWTSecret)
	}
}

func clearEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		prev, ok := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if ok {
				os.Setenv(key, prev)
			} else {
				os.Unsetenv(key)
			}
		})
	}
}
