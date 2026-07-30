# CLAUDE.md — Tutor Timetable

Instructions and context for Claude when working on this codebase.

## Linear

- Team: **Kubac**
- Project for all repo issues: **Tutor Timetable** — see [`docs/linear.md`](docs/linear.md)
- When creating issues via MCP (`save_issue`), set `project` to **Tutor Timetable** and `team` to **Kubac**
- Agentic workflow (isolated Docker sandbox): [`AGENTS.md`](AGENTS.md), [`docs/sandbox-workflow.md`](docs/sandbox-workflow.md), `./scripts/run-ticket.sh`

## Stack

- **Frontend**: SvelteKit 2, Svelte 5, Tailwind CSS, svelte-i18n, TypeScript
- **Backend**: Go 1.22, `chi` router, `pgx/v5` (PostgreSQL driver), JWT (httpOnly cookie)
- **Database**: PostgreSQL 16
- **Infra**: Docker Compose (services: `db`, `backend`, `frontend`)

## Running the project

**Full stack (Docker):**
```bash
docker compose up --build
```

**Backend only:**
```bash
cd backend && go mod tidy && go run .
# Requires DATABASE_URL env var pointing to a running Postgres instance
```

**Frontend only:**
```bash
cd frontend && npm install && npm run dev
# Vite proxies /auth /teacher /students /timetable → localhost:8081
```

**Type-check the frontend:**
```bash
cd frontend && npm run check
```

## Key conventions

### Backend (Go)

- All HTTP handlers live in `backend/api/` as methods on `*handlers`
- Error responses use error codes, never human-readable strings: `{"error": "NOT_FOUND"}`
- All DB queries use `pgx/v5` directly — no ORM
- Times are stored as PostgreSQL `TIME` and returned as `"HH:MM"` strings via `to_char(..., 'HH24:MI')`
- `day_of_week`: 0 = Monday, 6 = Sunday
- Auth is a JWT in an httpOnly cookie named `auth_token`; the `middleware.Auth` middleware injects `teacher_id` into context
- DB migrations are embedded in the binary via `go:embed`; add new `.sql` files to `backend/db/migrations/` with a numeric prefix (e.g. `002_add_column.sql`)
- New routes go in `backend/api/router.go` inside the protected group

### Frontend (Svelte/TypeScript)

- The app is a pure SPA (`ssr = false` in `+layout.ts`) — no server-side rendering in Phase 1
- All API calls go through `src/lib/api.ts` — do not use `fetch` directly in components
- All user-facing strings must use `$_('key')` from svelte-i18n — never hardcode display text
- Translation keys live in `src/lib/i18n/en.json`; always add a key there when adding new UI text
- Shared TypeScript types live in `src/lib/types.ts`
- Auth state is in `src/lib/stores/auth.ts` (`teacher` store, `authLoading` store)
- The auth guard is in `+layout.svelte` — unauthenticated users are redirected to `/login`
- Tailwind for all styling — no separate CSS files except `app.css` (which only imports Tailwind)
- `AvailabilityEditor.svelte` is the reusable click-and-drag availability grid; it takes a `slots` prop (array of `AvailabilitySlot`) and updates it in place

### i18n

- Phase 1 ships English only; the framework is in place for more locales
- Register new locales in `src/lib/i18n/index.ts` and add a corresponding JSON file
- Backend never returns human-readable error strings — only codes; the frontend translates them
- Date/time formatting should use locale-aware APIs (relevant for Phase 2)

### Scheduling

- Algorithm in `backend/scheduler/scheduler.go`
- Input: teacher settings, teacher availability slots, students + their availability, pinned lessons
- Output: new `[]models.Lesson` (unpinned only — caller deletes old unpinned and inserts these)
- Hard constraints: teacher/student availability, no overlaps, compulsory break slots
- Soft constraints: prefer early slots, penalise gaps > `max_gap_minutes`
- Slot granularity: 5 minutes throughout
- `day_of_week` 0–6, times as `"HH:MM"` strings

## Database schema (Phase 1)

```
teachers             id, email, password_hash, name, created_at
teacher_settings     teacher_id, working_start, working_end, lesson_duration_minutes,
                     max_gap_minutes, max_consecutive_lessons,
                     break_after_n_lessons, break_duration_minutes, locale
students             id, teacher_id, name, email, notes, created_at
availability_slots   id, owner_type ('teacher'|'student'), owner_id,
                     day_of_week, start_time, end_time
lessons              id, teacher_id, student_id, day_of_week,
                     start_time, end_time, is_pinned, created_at, updated_at
schema_migrations    filename, applied_at
```

## API routes

```
POST   /auth/login
POST   /auth/logout

GET    /teacher/me
GET    /teacher/settings
PUT    /teacher/settings
GET    /teacher/availability
PUT    /teacher/availability

GET    /students
POST   /students
GET    /students/:id
PUT    /students/:id
DELETE /students/:id
GET    /students/:id/availability
PUT    /students/:id/availability

GET    /timetable
POST   /timetable/generate
PUT    /timetable/lessons/:id
PATCH  /timetable/lessons/:id/pin
```

## Phase 2 planned features

(Not yet implemented — listed here to inform architecture decisions)

- Student share links for availability input
- Variable lesson lengths
- Joint/group lessons
- Multiple lessons per student
- Multi-teacher auth
- Availability levels (preferred / acceptable)
- Minimum availability enforcement
- Varying weekly schedules
- i18n: Czech, then German / French / Spanish
