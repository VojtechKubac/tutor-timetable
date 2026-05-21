#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <ticket-id> <short-description>"
  echo "Example: $0 kua-108 isolated-ticket-workflow"
  exit 1
fi

TICKET_ID_RAW="$1"
SLUG_RAW="$2"

normalize_slug() {
  printf '%s' "$1" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-{2,}/-/g'
}

TICKET_ID="$(normalize_slug "${TICKET_ID_RAW}")"
SLUG="$(normalize_slug "${SLUG_RAW}")"
BRANCH_NAME="${TICKET_ID}-${SLUG}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PARENT_DIR="$(dirname "${REPO_ROOT}")"
WORKTREE_DIR="${PARENT_DIR}/worktrees/${BRANCH_NAME}"

if [[ ! -d "${WORKTREE_DIR}" ]]; then
  if ! "${SCRIPT_DIR}/new-ticket-env.sh" "${TICKET_ID}" "${SLUG}"; then
    echo "Error: failed to create ticket environment for ${TICKET_ID}-${SLUG}" >&2
    exit 1
  fi
fi

cd "${WORKTREE_DIR}"
set -a
source .ticket-env
set +a

docker compose -f docker-compose.ticket.yml up -d --build

echo
echo "Ticket workflow ready:"
echo "  worktree: ${WORKTREE_DIR}"
echo "  branch:   ${BRANCH_NAME}"
echo
echo "Next step (inside container):"
echo "  docker compose -f docker-compose.ticket.yml exec ticket-dev bash"
echo
echo "Or from main clone:"
echo "  ./scripts/run-ticket.sh ${TICKET_ID}"
