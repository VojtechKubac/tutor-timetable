package api

import (
	"encoding/json"
	"net/http"

	"github.com/tutor-timetable/backend/middleware"
	"github.com/tutor-timetable/backend/models"
)

func (h *handlers) getMe(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	var t models.Teacher
	err := h.pool.QueryRow(r.Context(),
		"SELECT id, email, name, created_at FROM teachers WHERE id = $1",
		teacherID,
	).Scan(&t.ID, &t.Email, &t.Name, &t.CreatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *handlers) getSettings(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	var s models.TeacherSettings
	err := h.pool.QueryRow(r.Context(), `
		SELECT teacher_id,
		       to_char(working_start, 'HH24:MI'),
		       to_char(working_end,   'HH24:MI'),
		       lesson_duration_minutes, max_gap_minutes,
		       max_consecutive_lessons, break_after_n_lessons,
		       break_duration_minutes, locale
		FROM teacher_settings WHERE teacher_id = $1
	`, teacherID).Scan(
		&s.TeacherID, &s.WorkingStart, &s.WorkingEnd,
		&s.LessonDurationMinutes, &s.MaxGapMinutes,
		&s.MaxConsecutiveLessons, &s.BreakAfterNLessons,
		&s.BreakDurationMinutes, &s.Locale,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *handlers) updateSettings(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	var s models.TeacherSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}
	_, err := h.pool.Exec(r.Context(), `
		UPDATE teacher_settings SET
			working_start           = $1::time,
			working_end             = $2::time,
			lesson_duration_minutes = $3,
			max_gap_minutes         = $4,
			max_consecutive_lessons = $5,
			break_after_n_lessons   = $6,
			break_duration_minutes  = $7,
			locale                  = $8
		WHERE teacher_id = $9
	`,
		s.WorkingStart, s.WorkingEnd,
		s.LessonDurationMinutes, s.MaxGapMinutes,
		s.MaxConsecutiveLessons, s.BreakAfterNLessons,
		s.BreakDurationMinutes, s.Locale,
		teacherID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
