package scheduler

import (
	"testing"

	"github.com/tutor-timetable/backend/models"
)

func TestGenerateEmptyStudents(t *testing.T) {
	lessons := Generate(Input{
		Settings: models.TeacherSettings{
			LessonDurationMinutes: 45,
			WorkingStart:          "09:00",
			WorkingEnd:            "17:00",
		},
		TeacherSlots: []models.AvailabilitySlot{
			{DayOfWeek: 0, StartTime: "09:00", EndTime: "12:00"},
		},
		Students:        nil,
		StudentAvailMap: map[string][]models.AvailabilitySlot{},
	})
	if len(lessons) != 0 {
		t.Fatalf("expected no lessons, got %d", len(lessons))
	}
}

func TestGenerateSkipsStudentsWithoutAvailability(t *testing.T) {
	lessons := Generate(Input{
		Settings: models.TeacherSettings{
			LessonDurationMinutes: 45,
			WorkingStart:          "09:00",
			WorkingEnd:            "17:00",
		},
		TeacherSlots: []models.AvailabilitySlot{
			{DayOfWeek: 0, StartTime: "09:00", EndTime: "12:00"},
		},
		Students: []models.Student{
			{ID: "s1", Name: "Alice"},
		},
		StudentAvailMap: map[string][]models.AvailabilitySlot{
			"s1": {},
		},
	})
	if len(lessons) != 0 {
		t.Fatalf("expected skipped student, got %d lessons", len(lessons))
	}
}

func TestGeneratePlacesLessonOnOverlap(t *testing.T) {
	lessons := Generate(Input{
		Settings: models.TeacherSettings{
			LessonDurationMinutes: 45,
			WorkingStart:          "09:00",
			WorkingEnd:            "17:00",
		},
		TeacherSlots: []models.AvailabilitySlot{
			{DayOfWeek: 0, StartTime: "09:00", EndTime: "12:00"},
		},
		Students: []models.Student{
			{ID: "s1", Name: "Alice"},
		},
		StudentAvailMap: map[string][]models.AvailabilitySlot{
			"s1": {{DayOfWeek: 0, StartTime: "09:00", EndTime: "12:00"}},
		},
	})
	if len(lessons) != 1 {
		t.Fatalf("expected 1 lesson, got %d", len(lessons))
	}
	l := lessons[0]
	if l.StudentID != "s1" {
		t.Fatalf("StudentID: got %q", l.StudentID)
	}
	if l.DayOfWeek != 0 {
		t.Fatalf("DayOfWeek: got %d", l.DayOfWeek)
	}
	if l.StartTime != "09:00" {
		t.Fatalf("StartTime: got %q want 09:00", l.StartTime)
	}
	if l.EndTime != "09:45" {
		t.Fatalf("EndTime: got %q want 09:45", l.EndTime)
	}
	if l.IsPinned {
		t.Fatal("generated lesson must not be pinned")
	}
}
