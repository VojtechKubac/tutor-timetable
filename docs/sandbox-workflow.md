# Isolated Ticket Sandbox Workflow

How coding agents implement Linear tickets in fully isolated Docker sandboxes, with no access to the host filesystem.

## Why

The agent must be able to operate freely (edit anything, run anything, including Playwright E2E) without touching sensitive data on the host machine. The sandbox achieves this by:

- **No host filesystem mounts.** The workspace is a per-ticket Docker volume; the repo is cloned into it from GitHub.
- **No real secrets.** The container gets only the agent API key, a repo-scoped `GH_TOKEN`, and dev-only DB/JWT/seed values. The real `.env` is never copied; the sandbox `.env` is generated from the committed [`.env.sandbox`](../.env.sandbox) template.
- **Guardrails.** The git credential helper refuses `GH_TOKEN` for any repo other than `GH_ALLOWED_REPO_PATH`; a pre-push hook blocks pushes to `main`/`master`; the container drops all capabilities and has memory/CPU/pid limits.
- **Results leave only via GitHub.** The agent delivers a pushed branch + PR. The only host artifacts are logs under `~/.tutor-timetable/tickets/<ticket-id>/`.

## Prerequisites (host shell)

```bash
export GH_TOKEN=...          # repo-scoped GitHub token
export LINEAR_API_KEY=...    # to fetch the ticket
export ANTHROPIC_API_KEY=... # when AGENT_CLI=claude (default)
# or
export CURSOR_API_KEY=...    # with AGENT_CLI=cursor
```

## Usage

```bash
# Implement a ticket, fully hands-off
./scripts/run-ticket.sh kua-123

# Watch progress from another terminal
tail -f ~/.tutor-timetable/tickets/kua-123/agent.log

# Inspect or intervene manually (bash inside the sandbox)
./scripts/run-ticket.sh kua-123 --shell

# Re-run the agent in the same sandbox (e.g. after PR review comments)
./scripts/run-ticket.sh kua-123 --resume

# Tear down container, db and workspace volume (logs are kept)
./scripts/run-ticket.sh kua-123 --cleanup
```

Multiple tickets run in parallel — each gets its own compose project (`tt-<ticket-id>`) with its own db, container, and workspace volume.

An LLM running on the host triggers this the same way: it calls `./scripts/run-ticket.sh <ticket-id>` and monitors the log file. It never codes in the local clone (see the orchestrator role in [`AGENTS.md`](../AGENTS.md)).

## What happens on `run-ticket.sh <ticket-id>`

1. The ticket (title, description, URL) is fetched from Linear on the host.
2. `docker compose -p tt-<ticket-id> -f docker-compose.sandbox.yml up -d --build` starts a Postgres service and the agent container ([`Dockerfile.dev`](../Dockerfile.dev): Go, Node, gh, agent CLIs, Playwright system deps).
3. [`scripts/sandbox/init-workspace.sh`](../scripts/sandbox/init-workspace.sh) runs inside the container: clones the repo from GitHub into the `/workspace` volume, creates the ticket branch from `origin/main` (sticky across `--resume`, stored in `~/.tutor-timetable/tickets/<id>/branch`), and writes a dev-only `.env` from `.env.sandbox`.
4. The ticket prompt is written into the container and the agent CLI (Claude or Cursor, headless) runs to completion. Output streams to `~/.tutor-timetable/tickets/<id>/agent.log`.
5. The agent implements the ticket, runs quality checks and E2E, commits, pushes the branch, and opens a PR.

## Running the app + E2E inside the sandbox

There is no docker-in-docker; the app stack runs as plain processes inside the agent container, against the compose `db` service:

```bash
/workspace/scripts/sandbox/run-stack.sh start   # go build + run backend, production frontend build (adapter-node), waits until healthy
cd /workspace/e2e && npm install && npx playwright install chromium && npm test
/workspace/scripts/sandbox/run-stack.sh stop
```

Notes:

- Ports are auto-detected from `frontend/vite.config.ts` (backend from the proxy target, frontend from `server.port`), so the script works before and after the 3001/8081 port change lands on `main`.
- The frontend is served as a production build (`vite build` + `node build` with `PUBLIC_API_URL` set) instead of `vite dev` — no file watchers (the host inotify limit is shared with containers) and closer to the real deployment.
- Playwright browsers install into the shared `tutor-timetable-playwright-browsers` volume (`PLAYWRIGHT_BROWSERS_PATH=/ms-playwright`), so installs after the first are instant.
- Stack logs: `/tmp/tt-stack/backend.log`, `/tmp/tt-stack/frontend.log`.

## Volumes

| Volume | Scope | Removed by `--cleanup` |
|---|---|---|
| `tt-<ticket-id>_workspace` | per ticket | yes |
| `tutor-timetable-go-mod-cache` | shared | no |
| `tutor-timetable-npm-cache` | shared | no |
| `tutor-timetable-playwright-browsers` | shared | no |

The shared caches are external volumes created automatically by `run-ticket.sh`; delete them manually with `docker volume rm` if you ever want a cold start.

## Troubleshooting

- **Agent failed** — read `~/.tutor-timetable/tickets/<id>/agent.log`, then `--shell` into the sandbox to inspect `git status` and `/tmp/tt-stack/*.log`.
- **Clone fails** — check `GH_TOKEN` scope and `GH_ALLOWED_REPO_PATH` (defaults to `VojtechKubac/tutor-timetable.git`).
- **Stale workspace** — `--cleanup` and re-run; the workspace volume is disposable, anything pushed to GitHub survives.
