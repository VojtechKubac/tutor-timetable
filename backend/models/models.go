package models

import "time"

type Teacher struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TeacherSettings struct {
	TeacherID             string `json:"teacher_id"`
	WorkingStart          string `json:"working_start"`           // "HH:MM"
	WorkingEnd            string `json:"working_end"`             // "HH:MM"
	LessonDurationMinutes int    `json:"lesson_duration_minutes"` // 45
	MaxGapMinutes         int    `json:"max_gap_minutes"`
	MaxConsecutiveLessons int    `json:"max_consecutive_lessons"`
	BreakAfterNLessons    int    `json:"break_after_n_lessons"` // 0 = disabled
	BreakDurationMinutes  int    `json:"break_duration_minutes"`
	Locale                string `json:"locale"`
}

type Student struct {
	ID        string    `json:"id"`
	TeacherID string    `json:"teacher_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type AvailabilitySlot struct {
	ID        string `json:"id"`
	OwnerType string `json:"owner_type"` // "teacher" | "student"
	OwnerID   string `json:"owner_id"`
	DayOfWeek int    `json:"day_of_week"` // 0=Mon … 6=Sun
	StartTime string `json:"start_time"`  // "HH:MM"
	EndTime   string `json:"end_time"`    // "HH:MM"
}

type Lesson struct {
	ID          string    `json:"id"`
	TeacherID   string    `json:"teacher_id"`
	StudentID   string    `json:"student_id"`
	StudentName string    `json:"student_name,omitempty"`
	DayOfWeek   int       `json:"day_of_week"`
	StartTime   string    `json:"start_time"` // "HH:MM"
	EndTime     string    `json:"end_time"`   // "HH:MM"
	IsPinned    bool      `json:"is_pinned"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
