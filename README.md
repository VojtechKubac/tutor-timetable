# Tutor Timetable

A web application for music teachers to manage student availability and automatically generate optimised weekly lesson timetables.

## Stack

| Layer    | Technology                        |
|----------|-----------------------------------|
| Frontend | SvelteKit 2 + Tailwind CSS        |
| Backend  | Go 1.22 (`chi` router)            |
| Database | PostgreSQL 16                     |
| Infra    | Docker Compose                    |

## Quick start

```bash
cp .env.example .env
docker compose up --build
```

| Service  | URL                    |
|----------|------------------------|
| Frontend | http://localhost:3001  |
| Backend  | http://localhost:8081  |

Default login (created on first startup if no teachers exist):

```
Email:    teacher@example.com
Password: changeme
```

Change these in `.env` before first run.

## Local development (without Docker)

**Backend**
```bash
cd backend
go mod tidy
go run .
```
Requires a running PostgreSQL instance. Set `DATABASE_URL` in your environment or export it:
```bash
export DATABASE_URL="postgres://timetable:secret@localhost:5432/timetable?sslmode=disable"
```

**Frontend**
```bash
cd frontend
npm install
npm run dev
```
The Vite dev server proxies `/auth`, `/teacher`, `/students`, and `/timetable` to `http://localhost:8081` automatically — no CORS configuration needed in development.

## End-to-end tests (Playwright)

Against a running stack (`docker compose up`):

```bash
cd e2e
npm install
npx playwright install chromium   # once
npm test
```

See [`e2e/README.md`](e2e/README.md) for env overrides and coverage.

## Environment variables

Copy `.env.example` to `.env` and adjust as needed.

| Variable          | Default                  | Description                              |
|-------------------|--------------------------|------------------------------------------|
| `POSTGRES_DB`     | `timetable`              | Database name                            |
| `POSTGRES_USER`   | `timetable`              | Database user                            |
| `POSTGRES_PASSWORD` | `secret`               | Database password                        |
| `JWT_SECRET`      | `changeme-in-production` | Secret for signing auth tokens — **change this** |
| `FRONTEND_URL`    | `http://localhost:3001`  | Allowed CORS origin                      |
| `SEED_EMAIL`      | `teacher@example.com`    | Email for the seed teacher account       |
| `SEED_PASSWORD`   | `changeme`               | Password for the seed teacher account    |
| `SEED_NAME`       | `Music Teacher`          | Display name for the seed teacher        |
| `PUBLIC_API_URL`  | `http://localhost:8081`  | API base URL (used by the frontend container) |

## Features (Phase 1)

- **Teacher account** with login / logout
- **Student management** — add, edit, delete students
- **Availability editor** — click-and-drag weekly grid for teacher and each student
- **Timetable generation** — greedy + constraint-based scheduling; works with partial data
- **Pin lessons** — pinned lessons are kept fixed when regenerating the rest
- **Scheduling settings** — working hours, max gap, consecutive lessons, compulsory breaks
- **i18n-ready** — all strings in `en.json`; additional locales can be added

## Project structure

```
tutor-timetable/
├── docker-compose.yml
├── .env.example
├── e2e/                         # Playwright Phase 1 happy-path specs
├── backend/
│   ├── main.go                  # Entry point, DB connect, seed
│   ├── config/config.go         # Env-based config
│   ├── db/
│   │   ├── db.go                # Connection pool + migration runner
│   │   └── migrations/          # SQL migration files (embedded in binary)
│   ├── models/models.go         # Shared data types
│   ├── middleware/auth.go       # JWT cookie auth middleware
│   ├── api/                     # HTTP handlers (chi router)
│   └── scheduler/scheduler.go   # Greedy scheduling algorithm
└── frontend/
    └── src/
        ├── lib/
        │   ├── api.ts            # Typed API client
        │   ├── types.ts          # Shared TypeScript types
        │   ├── stores/auth.ts    # Auth state store
        │   ├── i18n/             # Translation files
        │   └── components/
        │       └── AvailabilityEditor.svelte
        └── routes/
            ├── +layout.svelte   # Nav sidebar + auth guard
            ├── +page.svelte     # Timetable view
            ├── login/
            ├── students/
            └── settings/
```

## Issue tracking

Linear project: [**Tutor Timetable**](https://linear.app/kuabc/project/tutor-timetable-6767972a978d) (team Kubac). Conventions for agents: [`docs/linear.md`](docs/linear.md).

## Isolated ticket workflow (agentic)

One Linear ticket → one git worktree/branch → one Docker container (same pattern as the trading repo). Full rules: [`AGENTS.md`](AGENTS.md).

```bash
# From main clone — orchestrator runs agent in container:
export GH_TOKEN=... LINEAR_API_KEY=...   # plus ANTHROPIC_API_KEY or CURSOR_API_KEY
./scripts/run-ticket.sh kua-108

# Or bootstrap manually:
./scripts/start-ticket-workflow.sh kua-108 short-slug
cd ../worktrees/kua-108-short-slug
set -a; source .ticket-env; set +a
docker compose -f docker-compose.ticket.yml exec ticket-dev bash
```

Ticket containers mount only the worktree at `/workspace` and use dev-only Postgres/JWT from `docker-compose.ticket.yml`.

## Planned (Phase 2)

- Student availability share links
- Variable lesson lengths
- Joint / group lessons
- Multiple lessons per student
- Multi-teacher accounts
- Availability levels ("preferred" vs "acceptable")
- Varying weekly schedules
- Additional languages (Czech, German, French, Spanish)
