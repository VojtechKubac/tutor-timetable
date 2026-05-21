package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tutor-timetable/backend/middleware"
	"github.com/tutor-timetable/backend/models"
)

func (h *handlers) listStudents(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	rows, err := h.pool.Query(r.Context(),
		`SELECT id, teacher_id, name, COALESCE(email,''), COALESCE(notes,''), created_at
		 FROM students WHERE teacher_id = $1 ORDER BY name`,
		teacherID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	defer rows.Close()

	students := []models.Student{}
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.TeacherID, &s.Name, &s.Email, &s.Notes, &s.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
		students = append(students, s)
	}
	writeJSON(w, http.StatusOK, students)
}

func (h *handlers) createStudent(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	var s models.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}
	if s.Name == "" {
		writeError(w, http.StatusBadRequest, "NAME_REQUIRED")
		return
	}
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO students (teacher_id, name, email, notes)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		teacherID, s.Name, s.Email, s.Notes,
	).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	s.TeacherID = teacherID
	writeJSON(w, http.StatusCreated, s)
}

func (h *handlers) getStudent(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	id := chi.URLParam(r, "id")
	var s models.Student
	err := h.pool.QueryRow(r.Context(),
		`SELECT id, teacher_id, name, COALESCE(email,''), COALESCE(notes,''), created_at
		 FROM students WHERE id = $1 AND teacher_id = $2`,
		id, teacherID,
	).Scan(&s.ID, &s.TeacherID, &s.Name, &s.Email, &s.Notes, &s.CreatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *handlers) updateStudent(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	id := chi.URLParam(r, "id")
	var s models.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}
	_, err := h.pool.Exec(r.Context(),
		"UPDATE students SET name=$1, email=$2, notes=$3 WHERE id=$4 AND teacher_id=$5",
		s.Name, s.Email, s.Notes, id, teacherID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) deleteStudent(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	id := chi.URLParam(r, "id")
	_, err := h.pool.Exec(r.Context(),
		"DELETE FROM students WHERE id=$1 AND teacher_id=$2",
		id, teacherID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
