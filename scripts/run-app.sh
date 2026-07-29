#!/usr/bin/env bash
# Start the full Tutor Timetable stack (db + backend + frontend) via Docker Compose.
#
# Usage:
#   ./scripts/run-app.sh          # foreground (Ctrl+C to stop)
#   ./scripts/run-app.sh -d       # detached
#   ./scripts/run-app.sh --down   # stop and remove containers

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

DETACHED=0
DOWN=0

for arg in "$@"; do
  case "${arg}" in
    -d|--detach)
      DETACHED=1
      ;;
    --down)
      DOWN=1
      ;;
    -h|--help)
      sed -n '2,8p' "$0" | sed 's/^# \?//'
      exit 0
      ;;
    *)
      echo "Unknown option: ${arg}" >&2
      echo "Usage: $0 [-d|--detach] [--down] [-h|--help]" >&2
      exit 1
      ;;
  esac
done

if ! command -v docker >/dev/null 2>&1; then
  echo "Error: docker is not installed or not on PATH." >&2
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "Error: docker compose is not available." >&2
  exit 1
fi

if [[ "${DOWN}" -eq 1 ]]; then
  echo "Stopping Tutor Timetable..."
  docker compose down
  exit 0
fi

if [[ ! -f .env ]]; then
  if [[ ! -f .env.example ]]; then
    echo "Error: .env.example not found; cannot create .env." >&2
    exit 1
  fi
  cp .env.example .env
  echo "Created .env from .env.example (edit secrets before production use)."
fi

echo "Building and starting Tutor Timetable..."
echo "  Frontend: http://localhost:3001"
echo "  Backend:  http://localhost:8081"
echo "  Login:    teacher@example.com / changeme  (override via .env)"
echo

if [[ "${DETACHED}" -eq 1 ]]; then
  docker compose up --build -d
  echo
  echo "Running in background. Stop with: ./scripts/run-app.sh --down"
else
  docker compose up --build
fi
