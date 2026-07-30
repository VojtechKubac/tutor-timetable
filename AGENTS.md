# AGENTS.md — Tutor Timetable (Primary Source of Truth)

Authoritative guide for coding agents working in this repository.

## Project Overview

Web app for music teachers: student availability, automatic weekly timetable generation, manual adjustments, pinned lessons. Stack: SvelteKit 2 SPA, Go 1.22 API (`chi`), PostgreSQL 16, Docker Compose. See [`PLAN.md`](PLAN.md) for Phase 1 scope and [`CLAUDE.md`](CLAUDE.md) for conventions.

## Repository Structure

```text
backend/                   Go API, scheduler, migrations (embedded)
frontend/                  SvelteKit app
docker-compose.yml         Full stack (db + backend + frontend)
docker-compose.sandbox.yml Isolated ticket sandbox (db + agent container, no host mounts)
scripts/                   Ticket sandbox workflow (run-ticket.sh, scripts/sandbox/)
docs/                      Agent and Linear docs
```

## Development Commands

```bash
# Full stack
docker compose up --build

# Backend only (needs Postgres + DATABASE_URL)
cd backend && go mod tidy && go test ./... && go run .

# Frontend only (Vite proxies API to :8080)
cd frontend && npm install && npm run dev && npm run check
```

## Conventions

- One Linear ticket = one PR; assign to Linear project **Tutor Timetable** (team **Kubac**). See [`docs/linear.md`](docs/linear.md).
- Backend errors: codes only (`{"error": "NOT_FOUND"}`), never human-readable strings.
- Frontend: all UI strings via `svelte-i18n`; API via `src/lib/api.ts` only.
- Branch from `main` only. Naming: `kua-{number}-short-description`.

## Workflow

- One Linear ticket = one branch + one PR, implemented inside an **isolated Docker sandbox** (no host filesystem mounts — the repo is cloned from GitHub into a per-ticket Docker volume).
- For agentic implementation, this workflow is the default unless explicitly overridden.
- **Always branch from `main`**, never from another feature branch.
- Full usage guide: [`docs/sandbox-workflow.md`](docs/sandbox-workflow.md).

### Orchestrator role — implementing a ticket from the main clone

When asked to implement a ticket from the **main** repository checkout (not inside a sandbox), **do not implement in the main clone**. Instead:

1. Run `./scripts/run-ticket.sh <ticket-id>` — fetches Linear, starts the sandbox (clone from GitHub into a Docker volume, dev-only `.env` from `.env.sandbox`), launches the in-container agent (Claude or Cursor).
2. Monitor output; on failure report `~/.tutor-timetable/tickets/<ticket-id>/agent.log`.
3. Only code in the main clone if the user explicitly asks or `run-ticket.sh` is unavailable.

```bash
# Host shell before run-ticket.sh:
#   GH_TOKEN, LINEAR_API_KEY — required
#   ANTHROPIC_API_KEY — when agent is Claude
#   CURSOR_API_KEY    — when agent is Cursor

./scripts/run-ticket.sh kua-108             # run agent
./scripts/run-ticket.sh kua-108 --resume    # re-run agent in existing sandbox
./scripts/run-ticket.sh kua-108 --shell     # interactive bash in the sandbox
./scripts/run-ticket.sh kua-108 --cleanup   # remove container, db, workspace volume
```

### Required Agent Preflight (in-container agents only)

Before coding inside a ticket sandbox:

```bash
pwd                          # must be /workspace
test -n "${TT_SANDBOX:-}"    # sandbox marker set by docker-compose.sandbox.yml
test -d /workspace/.git      # workspace was initialized (clone succeeded)
git branch --show-current    # must be the ticket branch (kua-*), never main
```

If any check fails, stop and report — do not proceed.

### Inside the sandbox

`ANTHROPIC_API_KEY`, `CURSOR_API_KEY`, and `GH_TOKEN` are forwarded from the host. Sandboxes use **dev-only** DB/JWT values from `docker-compose.sandbox.yml` and a dev-only `.env` generated from [`.env.sandbox`](.env.sandbox) — the maintainer's real `.env` never enters the sandbox.

`GH_ALLOWED_REPO_PATH` defaults to `VojtechKubac/tutor-timetable.git` (override for forks); the git credential helper refuses `GH_TOKEN` for any other repo, and pushes to `main`/`master` are blocked by a pre-push hook.

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
cd /workspace/frontend && npm install && npm run check
```

If `/workspace/e2e` exists, also run the Playwright E2E suite (the app stack runs as processes inside the sandbox container):

```bash
/workspace/scripts/sandbox/run-stack.sh start
cd /workspace/e2e && npm install && npx playwright install chromium && npm test
/workspace/scripts/sandbox/run-stack.sh stop
```

### Rules for agentic sessions

- Code only in `/workspace` inside the sandbox container.
- One sandbox (compose project `tt-<ticket-id>`) per ticket for parallel work.
- Deliver results only via `git push` + GitHub PR — nothing is written to the maintainer's machine.
- Clean up with `./scripts/run-ticket.sh <ticket-id> --cleanup` when done.
- **Never commit `.env`** with real secrets; `.env` is gitignored.
- Sandbox Postgres uses compose dev credentials only.

### Legacy worktree flow (manual use only)

`scripts/new-ticket-env.sh` and `scripts/start-ticket-workflow.sh` still create a host git worktree with a bind-mounted container (`docker-compose.ticket.yml`) for hands-on human work. Agents must use the sandbox flow above instead.

## What NOT to Do

- Never commit `.env` or production secrets.
- Do not use `fetch` directly in Svelte components — use `src/lib/api.ts`.
- Do not hardcode user-facing English strings in components.
