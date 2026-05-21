# Tutor Timetable Builder — Project Plan

## Overview

A web application for music teachers to manage student availability and automatically generate optimised weekly lesson timetables. Teachers can manually adjust generated timetables, pin lessons, and later share availability links with students.

---

## Tech Stack

| Layer      | Technology                          |
|------------|-------------------------------------|
| Frontend   | SvelteKit (SPA mode + SSR for share links) |
| Backend    | Go (`net/http` or `chi` router)     |
| Database   | PostgreSQL                          |
| Infra      | Docker Compose (3 services)         |

---

## Docker Compose Services

- **db** — PostgreSQL 16
- **backend** — Go API server
- **frontend** — SvelteKit (served via Node adapter or nginx)

---

## Phase 1

### Features

1. **Teacher account** (single account for Phase 1, multi-account architecture from the start)
2. **Student management** — CRUD for students
3. **Availability management**
   - Teacher sets their own weekly availability (5-minute granularity)
   - Teacher sets/edits availability on behalf of any student
4. **Timetable generation**
   - Fixed 45-minute lesson duration
   - Press "Generate" button — runs scheduling algorithm
   - Can generate with partial student data (not all students need availability filled in)
   - Pinned lessons are kept fixed during regeneration
5. **Manual adjustment** — drag-and-drop lessons on the timetable grid
6. **Pin lesson** — lock a lesson's slot so it survives regeneration
7. **Account settings** — scheduling preference defaults (configurable)

### Scheduling Algorithm

**Approach:** Greedy placement + local search (fast enough for ~20 students, feels instant)

1. Collect hard constraints (teacher available, student available, no overlaps, compulsory breaks)
2. Sort unpinned students by most-constrained availability first
3. Greedily assign each student to the earliest valid slot that minimises the cost function
4. Run a local search (swap/shift) pass to further reduce cost

**Hard constraints:**
- Lesson fits within teacher's available window
- Lesson fits within student's available window
- No two lessons overlap
- Compulsory break slots are respected

**Soft constraints (cost function, Phase 1 defaults — all configurable in settings):**
- Working window: 08:00–20:00
- Prefer earlier start times (linear penalty on start time)
- Penalise gaps > 30 minutes in the teacher's day
- Max consecutive lessons before a forced break: 4 (configurable)

**Granularity:** 5-minute slots throughout.

### Compulsory Breaks

- Configured per teacher in account settings
- Options:
  - Break after **every** lesson
  - Break after every **N** lessons
  - Break length: multiple of 5 minutes
- Break slots are treated as hard constraints during generation and manual placement

### Phase 1 Scheduling Defaults

| Setting                        | Default       |
|-------------------------------|---------------|
| Working window start           | 08:00         |
| Working window end             | 20:00         |
| Lesson duration                | 45 min        |
| Max gap before penalty         | 30 min        |
| Max consecutive lessons        | 4             |
| Compulsory break after N lessons | off (0)     |
| Compulsory break length        | 15 min        |

---

## Database Schema (Phase 1)

```sql
teachers
  id, email, password_hash, name, created_at

teacher_settings
  teacher_id (FK), working_start, working_end,
  max_gap_minutes, max_consecutive_lessons,
  break_after_n_lessons, break_duration_minutes

students
  id, teacher_id (FK), name, email, notes, created_at

availability_slots          -- shared table for both teacher and student
  id, owner_type (teacher|student), owner_id,
  day_of_week (0–6), start_time, end_time

lessons
  id, teacher_id (FK), student_id (FK),
  day_of_week, start_time, end_time,
  is_pinned, created_at, updated_at
```

---

## API Endpoints (Phase 1)

```
POST   /auth/login
POST   /auth/logout

GET    /teacher/me
PUT    /teacher/settings

GET    /students
POST   /students
GET    /students/:id
PUT    /students/:id
DELETE /students/:id

GET    /students/:id/availability
PUT    /students/:id/availability

GET    /teacher/availability
PUT    /teacher/availability

GET    /timetable
POST   /timetable/generate
PUT    /timetable/lessons/:id          -- move lesson (manual adjust)
PATCH  /timetable/lessons/:id/pin      -- toggle pin
```

---

## Frontend Pages (Phase 1)

| Route               | Description                                      |
|---------------------|--------------------------------------------------|
| `/login`            | Teacher login                                    |
| `/`                 | Timetable view — weekly grid, drag-and-drop      |
| `/students`         | Student list                                     |
| `/students/:id`     | Student detail + availability editor             |
| `/settings`         | Teacher availability + scheduling preferences    |

---

## Phase 2 Features (planned, not yet scoped)

- **Student availability link** — unique shareable URL per student; interactive slot picker + manual time entry
- **Variable lesson lengths** — per-student or per-lesson configuration
- **Joint lessons** — multiple students in one slot (e.g. chamber ensemble)
- **Multiple lessons per student** — e.g. instrument lesson + group lesson
- **Multi-teacher accounts** — full auth with per-teacher data isolation
- **Availability levels** — "fully available" vs "not ideal but possible" (soft constraint weighting)
- **Minimum available time enforcement** — student must fill in at least X hours of availability before submitting
- **Varying weekly schedules** — per-week availability overrides
- **Multi-language support** — English + Czech initially, German / French / Spanish later

### i18n Architecture Notes
- Phase 1 ships English only, but the frontend must be structured for i18n from day one:
  - All user-facing strings go through a translation layer (e.g. `svelte-i18n` or `paraglide-js`)
  - No hardcoded display strings in components
  - Locale stored in teacher settings (persisted in DB) and as a browser preference fallback
- Backend error messages and API responses should use error codes, not human-readable strings, so the frontend controls all displayed text
- Date/time formatting must use locale-aware formatting from the start (relevant for Czech vs English conventions)

---

## Project Directory Structure (proposed)

```
tutor-timetable/
├── docker-compose.yml
├── PLAN.md
├── backend/
│   ├── Dockerfile
│   ├── main.go
│   ├── api/
│   ├── db/
│   ├── scheduler/
│   └── models/
├── frontend/
│   ├── Dockerfile
│   ├── svelte.config.js
│   └── src/
│       ├── routes/
│       └── lib/
└── db/
    └── migrations/
```
