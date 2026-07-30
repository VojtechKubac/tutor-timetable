# AGENTS.md — Tutor Timetable (Primary Source of Truth)

Authoritative guide for coding agents working in this repository.

## Project Overview

Web app for music teachers: student availability, automatic weekly timetable generation, manual adjustments, pinned lessons. Stack: SvelteKit 2 SPA, Go 1.22 API (`chi`), PostgreSQL 16, Docker Compose. See [`PLAN.md`](PLAN.md) for Phase 1 scope and [`CLAUDE.md`](CLAUDE.md) for conventions.

## Repository Structure

```text
backend/           Go API, scheduler, migrations (embedded)
frontend/          SvelteKit app
docker-compose.yml Full stack (db + backend + frontend)
scripts/           Ticket worktree/container workflow
docs/              Agent and Linear docs
```

## Development Commands

```bash
# Full stack
docker compose up --build

# Backend only (needs Postgres + DATABASE_URL)
cd backend && go mod tidy && go test ./... && go run .

# Frontend only (Vite proxies API to :8081)
cd frontend && npm install && npm run dev

# Frontend validation (separate from the long-running dev server)
cd frontend && npm run check && npm test
```

## Conventions

- One Linear ticket = one PR; assign to Linear project **Tutor Timetable** (team **Kubac**). See [`docs/linear.md`](docs/linear.md).
- Backend errors: codes only (`{"error": "NOT_FOUND"}`), never human-readable strings.
- Frontend: all UI strings via `svelte-i18n`; API via `src/lib/api.ts` only.
- Branch from `main` only. Naming: `kua-{number}-short-description`.

## Workflow

- One Linear ticket = one dedicated git worktree + one branch + one Docker container.
- For agentic implementation, this workflow is the default unless explicitly overridden.
- **Always branch from `main`**, never from another feature branch.

### Orchestrator role — implementing a ticket from the main clone

When asked to implement a ticket from the **main** repository checkout (not a ticket worktree), **do not implement in the main clone**. Instead:

1. Run `./scripts/run-ticket.sh <ticket-id>` — fetches Linear, creates/reuses worktree, starts container, launches in-container agent (Claude or Cursor).
2. Monitor output; on failure report `.agent.log` under the worktree.
3. Only code in the main clone if the user explicitly asks or `run-ticket.sh` is unavailable.

```bash
# Host shell before run-ticket.sh:
#   GH_TOKEN, LINEAR_API_KEY — required
#   ANTHROPIC_API_KEY — when agent is Claude
#   CURSOR_API_KEY    — when agent is Cursor

./scripts/run-ticket.sh kua-108
```

### Required Agent Preflight (in-container agents only)

Before coding inside a ticket container:

```bash
pwd
test -f .ticket-env
set -a; source .ticket-env; set +a
docker compose -f docker-compose.ticket.yml ps
```

Path should be under `../worktrees/kua-*`. If not, stop and report — do not proceed.

### Ticket Environment Bootstrap

From the main repository checkout:

```bash
./scripts/start-ticket-workflow.sh kua-108 short-slug
```

Or manually:

```bash
./scripts/new-ticket-env.sh kua-108 short-slug
cd ../worktrees/kua-108-short-slug
set -a; source .ticket-env; set +a
docker compose -f docker-compose.ticket.yml up -d --build
docker compose -f docker-compose.ticket.yml exec ticket-dev bash
```

### Running the AI agent inside the container

`ANTHROPIC_API_KEY`, `CURSOR_API_KEY`, and `GH_TOKEN` are forwarded from the host. Ticket containers use **dev-only** DB/JWT values from `docker-compose.ticket.yml` — do not copy production `.env` into worktrees.

`GH_ALLOWED_REPO_PATH` defaults to `VojtechKubac/tutor-timetable.git` (override for forks).

**Claude Code:** `claude --dangerously-skip-permissions`  
**Cursor:** `cursor-agent -p --force --sandbox disabled "implement the ticket"`

Before `git push` / `gh pr create`:

```bash
gh auth status --hostname github.com
git ls-remote origin -h >/dev/null
```

### Quality checks (before commit)

```bash
cd /workspace/backend && go test ./...
cd /workspace/frontend && npm install && npm run check && npm test
```

### Rules for agentic sessions

- Code only in the ticket worktree mounted at `/workspace`.
- One worktree/container per ticket for parallel work.
- Stop containers when done.
- **Never commit `.env`** with real secrets; `.env` is gitignored.
- Ticket Postgres uses compose dev credentials only.

## What NOT to Do

- Never commit `.env` or production secrets.
- Do not use `fetch` directly in Svelte components — use `src/lib/api.ts`.
- Do not hardcode user-facing English strings in components.
