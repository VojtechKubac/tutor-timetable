package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tutor-timetable/backend/middleware"
	"github.com/tutor-timetable/backend/models"
	"github.com/tutor-timetable/backend/scheduler"
)

func (h *handlers) getTimetable(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	rows, err := h.pool.Query(r.Context(), `
		SELECT l.id, l.teacher_id, l.student_id, s.name,
		       l.day_of_week,
		       to_char(l.start_time, 'HH24:MI'),
		       to_char(l.end_time,   'HH24:MI'),
		       l.is_pinned, l.created_at, l.updated_at
		FROM lessons l
		JOIN students s ON s.id = l.student_id
		WHERE l.teacher_id = $1
		ORDER BY l.day_of_week, l.start_time
	`, teacherID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	defer rows.Close()

	lessons := []models.Lesson{}
	for rows.Next() {
		var l models.Lesson
		if err := rows.Scan(
			&l.ID, &l.TeacherID, &l.StudentID, &l.StudentName,
			&l.DayOfWeek, &l.StartTime, &l.EndTime,
			&l.IsPinned, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
		lessons = append(lessons, l)
	}
	writeJSON(w, http.StatusOK, lessons)
}

func (h *handlers) generateTimetable(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	ctx := r.Context()

	var settings models.TeacherSettings
	err := h.pool.QueryRow(ctx, `
		SELECT teacher_id,
		       to_char(working_start, 'HH24:MI'),
		       to_char(working_end,   'HH24:MI'),
		       lesson_duration_minutes, max_gap_minutes,
		       max_consecutive_lessons, break_after_n_lessons,
		       break_duration_minutes, locale
		FROM teacher_settings WHERE teacher_id = $1
	`, teacherID).Scan(
		&settings.TeacherID, &settings.WorkingStart, &settings.WorkingEnd,
		&settings.LessonDurationMinutes, &settings.MaxGapMinutes,
		&settings.MaxConsecutiveLessons, &settings.BreakAfterNLessons,
		&settings.BreakDurationMinutes, &settings.Locale,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	teacherSlots, err := h.LoadAvailability(ctx, "teacher", teacherID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	sRows, err := h.pool.Query(ctx,
		"SELECT id, name FROM students WHERE teacher_id=$1", teacherID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	var students []models.Student
	for sRows.Next() {
		var s models.Student
		_ = sRows.Scan(&s.ID, &s.Name)
		students = append(students, s)
	}
	sRows.Close()

	studentAvailMap := map[string][]models.AvailabilitySlot{}
	for _, s := range students {
		slots, err := h.LoadAvailability(ctx, "student", s.ID)
		if err == nil {
			studentAvailMap[s.ID] = slots
		}
	}

	pRows, err := h.pool.Query(ctx, `
		SELECT id, student_id, day_of_week,
		       to_char(start_time, 'HH24:MI'),
		       to_char(end_time,   'HH24:MI')
		FROM lessons WHERE teacher_id=$1 AND is_pinned=TRUE
	`, teacherID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	var pinned []models.Lesson
	for pRows.Next() {
		var l models.Lesson
		_ = pRows.Scan(&l.ID, &l.StudentID, &l.DayOfWeek, &l.StartTime, &l.EndTime)
		pinned = append(pinned, l)
	}
	pRows.Close()

	generated := scheduler.Generate(scheduler.Input{
		Settings:        settings,
		TeacherSlots:    teacherSlots,
		Students:        students,
		StudentAvailMap: studentAvailMap,
		PinnedLessons:   pinned,
	})

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		"DELETE FROM lessons WHERE teacher_id=$1 AND is_pinned=FALSE", teacherID,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	for _, l := range generated {
		if _, err := tx.Exec(ctx, `
			INSERT INTO lessons (teacher_id, student_id, day_of_week, start_time, end_time, is_pinned)
			VALUES ($1, $2, $3, $4::time, $5::time, FALSE)
		`, teacherID, l.StudentID, l.DayOfWeek, l.StartTime, l.EndTime); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.getTimetable(w, r)
}

func (h *handlers) moveLesson(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	id := chi.URLParam(r, "id")
	var req struct {
		DayOfWeek int    `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}
	_, err := h.pool.Exec(r.Context(), `
		UPDATE lessons
		SET day_of_week=$1, start_time=$2::time, end_time=$3::time, updated_at=NOW()
		WHERE id=$4 AND teacher_id=$5
	`, req.DayOfWeek, req.StartTime, req.EndTime, id, teacherID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) togglePin(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.TeacherID(r)
	id := chi.URLParam(r, "id")
	_, err := h.pool.Exec(r.Context(), `
		UPDATE lessons SET is_pinned=NOT is_pinned, updated_at=NOW()
		WHERE id=$1 AND teacher_id=$2
	`, id, teacherID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
