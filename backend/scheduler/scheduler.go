package scheduler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/tutor-timetable/backend/models"
)

type Input struct {
	Settings       models.TeacherSettings
	TeacherSlots   []models.AvailabilitySlot
	Students       []models.Student
	StudentAvailMap map[string][]models.AvailabilitySlot
	PinnedLessons  []models.Lesson
}

type timeSlot struct {
	day   int
	start int // minutes from midnight
}

// Generate returns new (unpinned) lesson placements.
// It uses a greedy approach: sort students by most-constrained first,
// then pick the earliest valid slot respecting all hard constraints.
func Generate(in Input) []models.Lesson {
	duration := in.Settings.LessonDurationMinutes
	if duration <= 0 {
		duration = 45
	}
	breakAfterN := in.Settings.BreakAfterNLessons
	breakDur := in.Settings.BreakDurationMinutes

	// Build occupied map from pinned lessons
	occupied := map[timeSlot]bool{}
	pinnedStudents := map[string]bool{}
	for _, l := range in.PinnedLessons {
		start := parseTime(l.StartTime)
		for t := start; t < start+duration; t += 5 {
			occupied[timeSlot{l.DayOfWeek, t}] = true
		}
		pinnedStudents[l.StudentID] = true
	}

	// Track lessons placed per day (for compulsory break counting)
	lessonsPerDay := map[int]int{}
	for _, l := range in.PinnedLessons {
		lessonsPerDay[l.DayOfWeek]++
	}

	// Sort students: most constrained (least total availability) first
	students := make([]models.Student, len(in.Students))
	copy(students, in.Students)
	sort.Slice(students, func(i, j int) bool {
		ai := totalAvailMinutes(in.StudentAvailMap[students[i].ID])
		aj := totalAvailMinutes(in.StudentAvailMap[students[j].ID])
		return ai < aj
	})

	var result []models.Lesson

	for _, student := range students {
		if pinnedStudents[student.ID] {
			continue
		}
		studentSlots := in.StudentAvailMap[student.ID]
		if len(studentSlots) == 0 {
			continue // no availability — skip
		}

		placed := false
		for day := 0; day <= 6 && !placed; day++ {
			teacherDay := slotsForDay(in.TeacherSlots, day)
			studentDay := slotsForDay(studentSlots, day)
			if len(teacherDay) == 0 || len(studentDay) == 0 {
				continue
			}

			for _, ts := range teacherDay {
				tsStart := parseTime(ts.StartTime)
				tsEnd := parseTime(ts.EndTime)

				for t := tsStart; t+duration <= tsEnd; t += 5 {
					if !coversSlot(studentDay, t, t+duration) {
						continue
					}
					if isOccupied(occupied, day, t, t+duration) {
						continue
					}

					// Place lesson
					result = append(result, models.Lesson{
						StudentID: student.ID,
						DayOfWeek: day,
						StartTime: formatTime(t),
						EndTime:   formatTime(t + duration),
					})
					for tt := t; tt < t+duration; tt += 5 {
						occupied[timeSlot{day, tt}] = true
					}
					lessonsPerDay[day]++

					// Mark compulsory break if required
					if breakAfterN > 0 && breakDur > 0 && lessonsPerDay[day]%breakAfterN == 0 {
						breakEnd := t + duration + breakDur
						for tt := t + duration; tt < breakEnd; tt += 5 {
							occupied[timeSlot{day, tt}] = true
						}
					}

					placed = true
					break
				}
				if placed {
					break
				}
			}
		}
	}

	return result
}

// --- helpers ---

func parseTime(s string) int {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h*60 + m
}

func formatTime(mins int) string {
	return fmt.Sprintf("%02d:%02d", mins/60, mins%60)
}

func slotsForDay(slots []models.AvailabilitySlot, day int) []models.AvailabilitySlot {
	var out []models.AvailabilitySlot
	for _, s := range slots {
		if s.DayOfWeek == day {
			out = append(out, s)
		}
	}
	return out
}

func coversSlot(slots []models.AvailabilitySlot, start, end int) bool {
	for _, s := range slots {
		if parseTime(s.StartTime) <= start && parseTime(s.EndTime) >= end {
			return true
		}
	}
	return false
}

func isOccupied(occupied map[timeSlot]bool, day, start, end int) bool {
	for t := start; t < end; t += 5 {
		if occupied[timeSlot{day, t}] {
			return true
		}
	}
	return false
}

func totalAvailMinutes(slots []models.AvailabilitySlot) int {
	total := 0
	for _, s := range slots {
		total += parseTime(s.EndTime) - parseTime(s.StartTime)
	}
	return total
}
