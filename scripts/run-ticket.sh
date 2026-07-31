#!/usr/bin/env bash
# run-ticket.sh — implement a Linear ticket inside an isolated Docker sandbox.
#
# The sandbox has NO host filesystem mounts: the repo is cloned from GitHub
# into a per-ticket Docker volume, a dev-only .env is generated from the
# committed .env.sandbox template, and the agent pushes results back as a
# branch + PR. The only host artifacts are logs under
# ~/.tutor-timetable/tickets/<ticket-id>/.
#
# Usage:
#   ./scripts/run-ticket.sh <ticket-id>            # create sandbox + run agent
#   ./scripts/run-ticket.sh <ticket-id> --resume   # re-run agent in existing sandbox
#   ./scripts/run-ticket.sh <ticket-id> --shell    # interactive bash in the sandbox
#   ./scripts/run-ticket.sh <ticket-id> --cleanup  # remove container, db and workspace volume
#
# Required host env (run/resume) — set in .env.agent (see .env.agent.example)
#   GH_TOKEN, LINEAR_API_KEY
#   Claude: CLAUDE_CODE_OAUTH_TOKEN (Pro/Max) or ANTHROPIC_API_KEY (API billing)
#   Cursor: CURSOR_API_KEY
# Optional: AGENT_CLI=claude|cursor, GH_ALLOWED_REPO_PATH=<owner>/<repo>.git

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Host secrets: copy .env.agent.example → .env.agent (gitignored).
if [[ -f "${REPO_ROOT}/.env.agent" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${REPO_ROOT}/.env.agent"
  set +a
fi

usage() {
  echo "Usage: $0 <ticket-id> [--resume|--shell|--cleanup]" >&2
  echo "Example: $0 kua-108" >&2
  exit 1
}

[[ $# -ge 1 ]] || usage

TICKET_ID="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
shift

if [[ ! "${TICKET_ID}" =~ ^[a-z]+-[0-9]+$ ]]; then
  echo "Error: ticket-id must match '<team>-<number>' (example: kua-108)." >&2
  exit 1
fi

MODE="run"
for arg in "$@"; do
  case "${arg}" in
    --resume)  MODE="resume" ;;
    --shell)   MODE="shell" ;;
    --cleanup) MODE="cleanup" ;;
    *) usage ;;
  esac
done

TICKET_ID_UPPER="$(printf '%s' "${TICKET_ID}" | tr '[:lower:]' '[:upper:]')"
COMPOSE_FILE="${REPO_ROOT}/docker-compose.sandbox.yml"
PROJECT="tt-${TICKET_ID}"
STATE_DIR="${HOME}/.tutor-timetable/tickets/${TICKET_ID}"

export HOST_UID="$(id -u)"
export HOST_GID="$(id -g)"
export GH_ALLOWED_REPO_PATH="${GH_ALLOWED_REPO_PATH:-VojtechKubac/tutor-timetable.git}"

compose() {
  docker compose -p "${PROJECT}" -f "${COMPOSE_FILE}" "$@"
}

if [[ "${MODE}" == "cleanup" ]]; then
  echo "Removing sandbox for ${TICKET_ID_UPPER} (container, db, workspace volume)..."
  compose down --volumes --remove-orphans
  echo "Done. Logs kept in ${STATE_DIR}. Shared caches are untouched."
  exit 0
fi

# Shared caches survive `down --volumes` because they are external volumes.
for vol in tutor-timetable-go-mod-cache tutor-timetable-npm-cache tutor-timetable-playwright-browsers; do
  docker volume create "${vol}" >/dev/null
done

if [[ "${MODE}" == "shell" ]]; then
  compose up -d --build
  echo "Opening shell in ${TICKET_ID_UPPER} sandbox (workspace may be empty if the agent never ran)..."
  exec docker compose -p "${PROJECT}" -f "${COMPOSE_FILE}" exec ticket-dev bash
fi

# --- run / resume ---

detect_agent_cli() {
  local override="${AGENT_CLI:-}"
  if [[ -n "${override}" ]]; then
    override="$(printf '%s' "${override}" | tr '[:upper:]' '[:lower:]')"
    if [[ "${override}" != "claude" && "${override}" != "cursor" ]]; then
      echo "Error: AGENT_CLI must be either 'claude' or 'cursor'." >&2
      exit 1
    fi
    printf '%s' "${override}"
    return
  fi

  if [[ "${CURSOR_AGENT:-}" == "1" || "${CURSOR_INVOKED_AS:-}" == "agent" ]]; then
    printf '%s' "cursor"
    return
  fi

  printf '%s' "claude"
}

AGENT_CLI="$(detect_agent_cli)"

for var in GH_TOKEN LINEAR_API_KEY; do
  if [[ -z "${!var:-}" ]]; then
    echo "Error: ${var} is not set in the host shell." >&2
    exit 1
  fi
done

if [[ "${AGENT_CLI}" == "cursor" ]]; then
  AGENT_LABEL="Cursor"
  AGENT_FOOTER_LINE="🤖 Generated with [Cursor](https://cursor.com)"
  if [[ -z "${CURSOR_API_KEY:-}" ]]; then
    echo "Error: CURSOR_API_KEY is required when AGENT_CLI=cursor." >&2
    exit 1
  fi
else
  AGENT_LABEL="Claude"
  AGENT_FOOTER_LINE="🤖 Generated with [Claude Code](https://claude.com/claude-code)"
  if [[ -z "${CLAUDE_CODE_OAUTH_TOKEN:-}" && -z "${ANTHROPIC_API_KEY:-}" ]]; then
    echo "Error: set CLAUDE_CODE_OAUTH_TOKEN (Pro/Max subscription) or ANTHROPIC_API_KEY (API billing) for Claude." >&2
    exit 1
  fi
  if [[ -n "${CLAUDE_CODE_OAUTH_TOKEN:-}" && -n "${ANTHROPIC_API_KEY:-}" ]]; then
    echo "Warning: both CLAUDE_CODE_OAUTH_TOKEN and ANTHROPIC_API_KEY are set; Claude Code will use the API key (pay-as-you-go)." >&2
  fi
fi

normalize_slug() {
  printf '%s' "$1" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//'
}

TEAM_KEY="$(printf '%s' "${TICKET_ID}" | sed 's/-[0-9]*$//' | tr '[:lower:]' '[:upper:]')"
TICKET_NUM="$(printf '%s' "${TICKET_ID}" | grep -oE '[0-9]+$')"

echo "Fetching ${TICKET_ID_UPPER} from Linear..."

LINEAR_RESPONSE=$(curl -s --connect-timeout 10 --max-time 30 -X POST https://api.linear.app/graphql \
  -H "Authorization: ${LINEAR_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"query\": \"{ issues(filter: { team: { key: { eq: \\\"${TEAM_KEY}\\\" } }, number: { eq: ${TICKET_NUM} } }, first: 1) { nodes { identifier title description url } } }\"}" \
  || { echo "Error: Linear API request failed" >&2; exit 1; })

eval "$(
printf '%s' "${LINEAR_RESPONSE}" | python3 -c '
import json, shlex, sys
try:
    d = json.loads(sys.stdin.read())
    nodes = d.get("data", {}).get("issues", {}).get("nodes", [])
    n = nodes[0] if nodes else {}
    print("TICKET_TITLE=" + shlex.quote(n.get("title") or ""))
    print("TICKET_DESC=" + shlex.quote(n.get("description") or ""))
    print("TICKET_URL=" + shlex.quote(n.get("url") or ""))
except Exception as e:
    print("echo " + shlex.quote("JSON parse error: " + str(e)) + " >&2; exit 1")
'
)"

if [[ -z "${TICKET_TITLE}" ]]; then
  echo "Error: could not fetch ${TICKET_ID_UPPER} from Linear. Check LINEAR_API_KEY and ticket ID." >&2
  exit 1
fi

echo "  Title: ${TICKET_TITLE}"
echo "  URL:   ${TICKET_URL}"

mkdir -p "${STATE_DIR}"

# Branch name is sticky per ticket so --resume reuses the same branch even if
# the Linear title changed.
BRANCH_FILE="${STATE_DIR}/branch"
if [[ -f "${BRANCH_FILE}" ]]; then
  BRANCH_NAME="$(cat "${BRANCH_FILE}")"
else
  SLUG="$(normalize_slug "${TICKET_TITLE}" | cut -c1-40 | sed 's/-$//')"
  BRANCH_NAME="${TICKET_ID}-${SLUG}"
  printf '%s\n' "${BRANCH_NAME}" > "${BRANCH_FILE}"
fi

echo "  Branch: ${BRANCH_NAME}"
echo "  Sandbox project: ${PROJECT}"

if ! compose up -d --build; then
  echo "Error: failed to start sandbox for ${TICKET_ID_UPPER}" >&2
  exit 1
fi

REPO_URL="https://github.com/${GH_ALLOWED_REPO_PATH%.git}.git"

echo "Initializing workspace (clone from ${REPO_URL})..."
if ! compose exec -T ticket-dev bash -s -- "${REPO_URL}" "${BRANCH_NAME}" \
    < "${SCRIPT_DIR}/sandbox/init-workspace.sh"; then
  echo "Error: workspace initialization failed for ${TICKET_ID_UPPER}" >&2
  exit 1
fi

PROMPT_FILE="${STATE_DIR}/prompt.txt"
LOG_FILE="${STATE_DIR}/agent.log"

RESUME_NOTE=""
if [[ "${MODE}" == "resume" ]]; then
  RESUME_NOTE="
## Resumed session

This sandbox already contains earlier work on this ticket. Start by inspecting
\`git status\`, \`git log origin/main..HEAD\` and any open PR for branch
${BRANCH_NAME} (\`gh pr list --head ${BRANCH_NAME}\`), then continue — e.g.
address review comments or finish incomplete work.
"
fi

cat > "${PROMPT_FILE}" <<PROMPT
You are a coding agent implementing Linear ticket ${TICKET_ID_UPPER}.
You run inside an isolated sandbox container; the repo is cloned at /workspace
on branch ${BRANCH_NAME}. There is no access to the maintainer's machine —
deliver results only via git push and a GitHub PR.
${RESUME_NOTE}
## Ticket

**Title:** ${TICKET_TITLE}
**URL:** ${TICKET_URL}

**Description:**

${TICKET_DESC}

## Instructions

1. Read AGENTS.md for all project conventions before making any changes.
2. Implement the ticket as described above.
3. Before committing, run quality checks and fix failures:
   \`\`\`
   cd /workspace/backend && go test ./...
   cd /workspace/frontend && npm install && npm run check && npm test
   \`\`\`
4. If /workspace/e2e exists, also run the Playwright E2E suite and fix failures:
   \`\`\`
   /workspace/scripts/sandbox/run-stack.sh start
   cd /workspace/e2e && npm install && npx playwright install chromium && npm test
   /workspace/scripts/sandbox/run-stack.sh stop
   \`\`\`
   (Browsers install into a cached volume, so this is fast after the first run.)
5. Commit: "${TICKET_ID_UPPER}: <short description>"
6. Push: \`git push -u origin ${BRANCH_NAME}\`
7. Open PR with Summary, Test plan checklist, and footer:
   ${AGENT_FOOTER_LINE}
8. If CodeRabbit is enabled, address all review comments before finishing.

Work autonomously to completion. Do not pause for confirmation.
PROMPT

compose exec -T ticket-dev bash -c 'cat > /workspace/.agent-prompt.txt' < "${PROMPT_FILE}"

echo ""
echo "Launching ${AGENT_LABEL} agent for ${TICKET_ID_UPPER} (log: ${LOG_FILE})..."
echo ""

set +e
if [[ "${AGENT_CLI}" == "cursor" ]]; then
  compose exec -T ticket-dev \
    bash -c 'cursor-agent -p --force --sandbox disabled "$(cat /workspace/.agent-prompt.txt)"' \
    2>&1 | tee "${LOG_FILE}"
  AGENT_EXIT=${PIPESTATUS[0]}
else
  compose exec -T ticket-dev \
    bash -c 'claude --dangerously-skip-permissions -p "$(cat /workspace/.agent-prompt.txt)"' \
    2>&1 | tee "${LOG_FILE}"
  AGENT_EXIT=${PIPESTATUS[0]}
fi
set -e

compose exec -T ticket-dev bash -c 'rm -f /workspace/.agent-prompt.txt' || true

echo ""
if [[ ${AGENT_EXIT} -eq 0 ]]; then
  echo "✓ ${TICKET_ID_UPPER} completed — see ${LOG_FILE}"
  echo "  Inspect:  $0 ${TICKET_ID} --shell"
  echo "  Clean up: $0 ${TICKET_ID} --cleanup"
else
  echo "✗ ${TICKET_ID_UPPER} failed (exit ${AGENT_EXIT}) — see ${LOG_FILE}" >&2
  exit "${AGENT_EXIT}"
fi
