# Scripts — ticket workflow notes

Practical notes for the ticket automation scripts. The full sandbox architecture
is documented in [`docs/sandbox-workflow.md`](../docs/sandbox-workflow.md).

## Which script does what

| Script | What it does | Launches the agent? |
|---|---|---|
| `run-ticket.sh` | Fetches the ticket from Linear, builds an isolated sandbox (repo cloned into a Docker volume, no host mounts), runs the agent headlessly to completion | **Yes** |
| `start-ticket-workflow.sh` | Legacy worktree flow: creates a git worktree + `.ticket-env`, starts an idle container | No — bootstrap only |
| `new-ticket-env.sh` | Creates the worktree and `.ticket-env` (called by `start-ticket-workflow.sh`) | No |
| `sandbox/init-workspace.sh` | Runs inside the sandbox: clones the repo, creates the ticket branch, writes dev `.env` | — |
| `sandbox/run-stack.sh` | Runs backend + frontend inside the sandbox for E2E tests | — |

If you ran `start-ticket-workflow.sh` and only see an idle container: that is
expected — nothing implements the ticket until `run-ticket.sh` runs. The
sandbox flow (`run-ticket.sh`) does **not** use those worktrees at all; it
clones fresh from GitHub into a per-ticket volume (compose project
`tt-<ticket-id>`).

## Getting the tokens (`.env.agent`)

Copy `.env.agent.example` to `.env.agent` (gitignored) and fill in:

- **`GH_TOKEN`** — GitHub token used for clone/push/PR. `gh auth token` works,
  but a **fine-grained PAT** (github.com → Settings → Developer settings →
  Fine-grained tokens) scoped to this one repo with *Contents* and *Pull
  requests* read/write is safer — see the `GH_ALLOWED_REPO_PATH` caveat below.
- **`LINEAR_API_KEY`** — Linear **personal API key**, created at linear.app →
  Settings → *Security & access* → *Personal API keys*. It must start with
  `lin_api_`. The Linear connection inside Cursor/Claude (MCP OAuth) is not
  reusable here. A wrong value fails as
  `Error: could not fetch ... from Linear` (HTTP 401).
- **`CLAUDE_CODE_OAUTH_TOKEN`** (default agent) — run `claude setup-token`
  once on the host, paste the printed `sk-ant-...` token. Bills against your
  Pro/Max subscription. Alternative: `ANTHROPIC_API_KEY` (pay-as-you-go); do
  not set both, the API key wins.
- **`CURSOR_API_KEY`** — only when `AGENT_CLI=cursor`. Created at cursor.com →
  Dashboard → Integrations → API Keys; your IDE login is not exportable.

Sanity check after editing — every key has a distinct, correct prefix
(`gho_`/`github_pat_`, `lin_api_`, `sk-ant-`). A classic failure mode is
pasting the same clipboard value into multiple variables.

## Role of `GH_ALLOWED_REPO_PATH`

Blast-radius limiter baked into the container image (`Dockerfile.dev`). A git
credential helper hands `GH_TOKEN` out **only** for HTTPS operations against
`github.com/<GH_ALLOWED_REPO_PATH>` (default:
`VojtechKubac/tutor-timetable.git`); any other repo gets
"Refusing GH_TOKEN for ...". Override it only when working on a fork.

Caveat: this guards **git** operations only. The `gh` CLI reads `GH_TOKEN`
from the environment directly and bypasses the credential helper, so `gh api`
inside the sandbox can reach anything the token can. That is why a repo-scoped
fine-grained PAT is preferred over a broad token.

## Is the agent actually running?

`claude -p` (headless print mode) buffers its output and writes it **at the
end of the run**, so a quiet `~/.tutor-timetable/tickets/<id>/agent.log` does
not mean the agent is stuck. To check:

```bash
# Live process? Expect a `claude` (or `cursor-agent`) process using CPU
docker compose -p tt-<ticket-id> -f docker-compose.sandbox.yml exec ticket-dev ps aux

# Progress? Files start showing up as modified after the initial read phase
docker compose -p tt-<ticket-id> -f docker-compose.sandbox.yml exec ticket-dev \
  bash -c 'cd /workspace && git status --short && git log --oneline -3'
```

When the run finishes, `run-ticket.sh` prints `✓ <TICKET> completed` and the
full transcript lands in the log file.

Harmless startup messages you can ignore:

- `Warning: no stdin data received in 3s` — fixed by redirecting stdin (`< /dev/null`)
  in `run-ticket.sh`; harmless if you still see it on an older checkout.

## Troubleshooting

### `service "ticket-dev" is not running`

The sandbox container stopped (often exit 137 after Docker Desktop restart, manual
`docker stop`, or an idle `bash` PID 1 exiting). The workspace volume and any
partial git work are still there.

```bash
# Restart the sandbox and resume the agent (keeps /workspace volume)
./scripts/run-ticket.sh kua-96 --resume

# Or inspect partial work first
./scripts/run-ticket.sh kua-96 --shell
cd /workspace && git status && git diff --stat
```

If `--resume` keeps failing, recreate the container (workspace volume is kept):

```bash
docker compose -p tt-kua-96 -f docker-compose.sandbox.yml up -d --force-recreate
./scripts/run-ticket.sh kua-96 --resume
```

### Agent log is empty or very short

`claude -p` buffers output until the run finishes — a quiet log during a long
run is normal. Check the process is alive:

```bash
docker compose -p tt-kua-96 -f docker-compose.sandbox.yml exec ticket-dev ps aux | grep claude
```
