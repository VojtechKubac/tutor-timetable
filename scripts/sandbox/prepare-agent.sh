#!/usr/bin/env bash
# prepare-agent.sh — one-time per-container setup before launching the agent
# (runs INSIDE the ticket container; scripts/run-ticket.sh pipes it in).
#
# - Marks /workspace as trusted for Claude Code (silences trust-dialog noise)
# - Verifies agent credentials were forwarded into the container

set -euo pipefail

AGENT_CLI="${1:-claude}"

if [[ "${AGENT_CLI}" == "claude" ]]; then
  python3 <<'PY'
import json, os
path = os.path.expanduser("~/.claude.json")
data = {}
if os.path.isfile(path):
    with open(path) as f:
        data = json.load(f)
data.setdefault("projects", {})["/workspace"] = {"hasTrustDialogAccepted": True}
with open(path, "w") as f:
    json.dump(data, f, indent=2)
    f.write("\n")
os.chmod(path, 0o600)
print("Claude Code: /workspace marked as trusted")
PY

  if [[ -z "${CLAUDE_CODE_OAUTH_TOKEN:-}" && -z "${ANTHROPIC_API_KEY:-}" ]]; then
    echo "Error: no Claude credentials in the container." >&2
    echo "  Set CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_API_KEY in .env.agent, then recreate:" >&2
    echo "  docker compose -p tt-<ticket> -f docker-compose.sandbox.yml up -d --force-recreate" >&2
    exit 1
  fi
  echo "Claude Code: credentials present in container"
elif [[ "${AGENT_CLI}" == "cursor" ]]; then
  if [[ -z "${CURSOR_API_KEY:-}" ]]; then
    echo "Error: CURSOR_API_KEY is not set in the container environment." >&2
    exit 1
  fi
  echo "Cursor: CURSOR_API_KEY is set"
else
  echo "Error: unknown agent CLI: ${AGENT_CLI}" >&2
  exit 1
fi
