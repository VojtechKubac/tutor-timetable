package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tutor-timetable/backend/middleware"
	"github.com/tutor-timetable/backend/models"
)

func (h *handlers) getTeacherAvailability(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	h.fetchAvailability(w, r.Context(), "teacher", teacherID)
}

func (h *handlers) updateTeacherAvailability(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	h.replaceAvailability(w, r, "teacher", teacherID)
}

func (h *handlers) getStudentAvailability(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	studentID := chi.URLParam(r, "id")
	if !h.studentBelongsToTeacher(r.Context(), studentID, teacherID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}
	h.fetchAvailability(w, r.Context(), "student", studentID)
}

func (h *handlers) updateStudentAvailability(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	studentID := chi.URLParam(r, "id")
	if !h.studentBelongsToTeacher(r.Context(), studentID, teacherID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}
	h.replaceAvailability(w, r, "student", studentID)
}

func (h *handlers) fetchAvailability(w http.ResponseWriter, ctx context.Context, ownerType, ownerID string) {
	rows, err := h.pool.Query(ctx, `
		SELECT id, owner_type, owner_id, day_of_week,
		       to_char(start_time, 'HH24:MI'),
		       to_char(end_time,   'HH24:MI')
		FROM availability_slots
		WHERE owner_type=$1 AND owner_id=$2
		ORDER BY day_of_week, start_time
	`, ownerType, ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	defer rows.Close()

	slots := []models.AvailabilitySlot{}
	for rows.Next() {
		var s models.AvailabilitySlot
		if err := rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.DayOfWeek, &s.StartTime, &s.EndTime); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
		slots = append(slots, s)
	}
	writeJSON(w, http.StatusOK, slots)
}

func (h *handlers) replaceAvailability(w http.ResponseWriter, r *http.Request, ownerType, ownerID string) {
	var slots []models.AvailabilitySlot
	if err := json.NewDecoder(r.Body).Decode(&slots); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	defer tx.Rollback(r.Context())

	if _, err := tx.Exec(r.Context(),
		"DELETE FROM availability_slots WHERE owner_type=$1 AND owner_id=$2",
		ownerType, ownerID,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	for _, s := range slots {
		if _, err := tx.Exec(r.Context(), `
			INSERT INTO availability_slots (owner_type, owner_id, day_of_week, start_time, end_time)
			VALUES ($1, $2, $3, $4::time, $5::time)
		`, ownerType, ownerID, s.DayOfWeek, s.StartTime, s.EndTime); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) studentBelongsToTeacher(ctx context.Context, studentID, teacherID string) bool {
	var count int
	_ = h.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM students WHERE id=$1 AND teacher_id=$2",
		studentID, teacherID,
	).Scan(&count)
	return count > 0
}

// LoadAvailability is exported for use by the timetable handler.
func (h *handlers) LoadAvailability(ctx context.Context, ownerType, ownerID string) ([]models.AvailabilitySlot, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT id, owner_type, owner_id, day_of_week,
		       to_char(start_time, 'HH24:MI'),
		       to_char(end_time,   'HH24:MI')
		FROM availability_slots
		WHERE owner_type=$1 AND owner_id=$2
		ORDER BY day_of_week, start_time
	`, ownerType, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []models.AvailabilitySlot
	for rows.Next() {
		var s models.AvailabilitySlot
		if err := rows.Scan(&s.ID, &s.OwnerType, &s.OwnerID, &s.DayOfWeek, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, nil
}
