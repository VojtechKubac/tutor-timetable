package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tutor-timetable/backend/config"
	"github.com/tutor-timetable/backend/middleware"
)

type handlers struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

func NewRouter(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}))

	h := &handlers{pool: pool, cfg: cfg}

	// Public
	r.Post("/auth/login", h.login)
	r.Post("/auth/logout", h.logout)

	// Protected
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))

		r.Get("/teacher/me", h.getMe)
		r.Get("/teacher/settings", h.getSettings)
		r.Put("/teacher/settings", h.updateSettings)
		r.Get("/teacher/availability", h.getTeacherAvailability)
		r.Put("/teacher/availability", h.updateTeacherAvailability)

		r.Get("/students", h.listStudents)
		r.Post("/students", h.createStudent)
		r.Get("/students/{id}", h.getStudent)
		r.Put("/students/{id}", h.updateStudent)
		r.Delete("/students/{id}", h.deleteStudent)
		r.Get("/students/{id}/availability", h.getStudentAvailability)
		r.Put("/students/{id}/availability", h.updateStudentAvailability)

		r.Get("/timetable", h.getTimetable)
		r.Post("/timetable/generate", h.generateTimetable)
		r.Put("/timetable/lessons/{id}", h.moveLesson)
		r.Patch("/timetable/lessons/{id}/pin", h.togglePin)
	})

	return r
}
