# End-to-end tests (Playwright)

Phase 1 happy-path coverage for Tutor Timetable. Specs live in `e2e/specs/` and talk to a running full stack (Docker Compose or equivalent).

## Prerequisites

1. Copy env and start the app from the repo root:

```bash
cp .env.example .env   # if needed
docker compose up --build -d
```

Frontend: http://localhost:3001 · Backend: http://localhost:8081  
Seed login defaults: `SEED_EMAIL` / `SEED_PASSWORD` from `.env` (`teacher@example.com` / `changeme`).

2. Install Playwright (once per machine / CI image):

```bash
cd e2e
npm install
npx playwright install chromium
```

## Run

```bash
cd e2e
npm test
```

Optional:

| Command | Purpose |
|---------|---------|
| `npm run test:headed` | Show the browser |
| `npm run test:ui` | Playwright UI mode |
| `PLAYWRIGHT_BASE_URL=http://localhost:3001 npm test` | Override frontend URL |
| `PLAYWRIGHT_API_URL=http://localhost:8081 npm test` | Override API URL (API setup helpers) |

Credentials are read from the repo-root `.env` (`SEED_*`, `FRONTEND_URL`, `PUBLIC_API_URL`). Do not commit secrets.

## What is covered

1. Auth guard, login, logout  
2. Student create / edit / delete  
3. Teacher and student availability editors (save + reload)  
4. Timetable generate; pinned lesson preserved on regenerate  
5. Pin / unpin  
6. Settings persistence  

Drag-and-drop lesson moves are out of scope until that UI lands.
