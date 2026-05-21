package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tutor-timetable/backend/api"
	"github.com/tutor-timetable/backend/config"
	"github.com/tutor-timetable/backend/db"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(pool); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	if err := seedTeacher(pool, cfg); err != nil {
		log.Printf("warning: seed teacher: %v", err)
	}

	router := api.NewRouter(pool, cfg)

	addr := ":" + cfg.Port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func seedTeacher(pool *pgxpool.Pool, cfg *config.Config) error {
	ctx := context.Background()

	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM teachers").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.SeedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var teacherID string
	err = pool.QueryRow(ctx,
		"INSERT INTO teachers (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id",
		cfg.SeedEmail, string(hash), cfg.SeedName,
	).Scan(&teacherID)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx,
		"INSERT INTO teacher_settings (teacher_id) VALUES ($1)",
		teacherID,
	)
	if err != nil {
		return err
	}

	log.Printf("seeded teacher: %s", cfg.SeedEmail)
	return nil
}
