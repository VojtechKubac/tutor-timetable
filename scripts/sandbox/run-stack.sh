#!/usr/bin/env bash
# run-stack.sh — start/stop the full app stack as processes INSIDE the ticket
# sandbox container, so Playwright E2E tests can run without docker-in-docker.
#
# Usage (from /workspace):
#   scripts/sandbox/run-stack.sh start   # backend (go) + frontend (vite) + wait until healthy
#   scripts/sandbox/run-stack.sh stop
#   scripts/sandbox/run-stack.sh status
#
# Postgres comes from the compose `db` service (DATABASE_URL is already set in
# the container environment). Ports are read from frontend/vite.config.ts so
# the script works both before and after the port changes on kua-109
# (3001/8081) land on main.

set -euo pipefail

WORKSPACE=/workspace
RUN_DIR=/tmp/tt-stack
CMD="${1:-start}"

cd "${WORKSPACE}"
mkdir -p "${RUN_DIR}"

VITE_CONFIG=frontend/vite.config.ts
FRONTEND_PORT="${FRONTEND_PORT:-$(grep -oE 'port:[[:space:]]*[0-9]+' "${VITE_CONFIG}" 2>/dev/null | grep -oE '[0-9]+' | head -1 || true)}"
FRONTEND_PORT="${FRONTEND_PORT:-3001}"
BACKEND_PORT="${BACKEND_PORT:-$(grep -oE 'localhost:[0-9]+' "${VITE_CONFIG}" 2>/dev/null | grep -oE '[0-9]+' | head -1 || true)}"
BACKEND_PORT="${BACKEND_PORT:-8080}"

BACKEND_PGID_FILE="${RUN_DIR}/backend.pgid"
FRONTEND_PGID_FILE="${RUN_DIR}/frontend.pgid"

wait_for_http() {
  local url="$1" name="$2" timeout="${3:-60}" i=0
  until curl -s -o /dev/null --max-time 2 "${url}"; do
    i=$((i + 1))
    if [[ ${i} -ge ${timeout} ]]; then
      echo "Error: ${name} did not respond at ${url} within ${timeout}s" >&2
      return 1
    fi
    sleep 1
  done
  echo "${name} is up at ${url}"
}

kill_group() {
  local pgid_file="$1" name="$2"
  if [[ -f "${pgid_file}" ]]; then
    local pgid
    pgid="$(cat "${pgid_file}")"
    if kill -TERM -- "-${pgid}" 2>/dev/null; then
      echo "Stopped ${name} (pgid ${pgid})"
    fi
    rm -f "${pgid_file}"
  fi
}

case "${CMD}" in
  start)
    # Dev-only seed/config values from the sandbox .env; keep the compose
    # DATABASE_URL (it points at the db service).
    SAVED_DATABASE_URL="${DATABASE_URL:-}"
    if [[ -f .env ]]; then
      set -a
      # shellcheck disable=SC1091
      source .env
      set +a
    fi
    if [[ -n "${SAVED_DATABASE_URL}" ]]; then
      export DATABASE_URL="${SAVED_DATABASE_URL}"
    fi

    echo "Building backend..."
    (cd backend && go build -o "${RUN_DIR}/backend-bin" .)

    echo "Starting backend on :${BACKEND_PORT}..."
    PORT="${BACKEND_PORT}" FRONTEND_URL="http://localhost:${FRONTEND_PORT}" \
      setsid "${RUN_DIR}/backend-bin" > "${RUN_DIR}/backend.log" 2>&1 &
    echo $! > "${BACKEND_PGID_FILE}"
    wait_for_http "http://localhost:${BACKEND_PORT}/" backend 60 || {
      tail -50 "${RUN_DIR}/backend.log" >&2
      exit 1
    }

    if [[ ! -d frontend/node_modules ]]; then
      echo "Installing frontend dependencies..."
      (cd frontend && npm install)
    fi

    # Production build served by adapter-node instead of `vite dev`: no file
    # watchers (the host inotify limit is shared with containers) and closer
    # to the real deployment. The SPA reads PUBLIC_API_URL at runtime.
    echo "Building frontend..."
    (cd frontend && npm run build)

    echo "Starting frontend on :${FRONTEND_PORT}..."
    (cd frontend && \
      PORT="${FRONTEND_PORT}" HOST=0.0.0.0 PUBLIC_API_URL="http://localhost:${BACKEND_PORT}" \
      setsid node build > "${RUN_DIR}/frontend.log" 2>&1 &
      echo $! > "${FRONTEND_PGID_FILE}")
    wait_for_http "http://localhost:${FRONTEND_PORT}/" frontend 120 || {
      tail -50 "${RUN_DIR}/frontend.log" >&2
      exit 1
    }

    echo
    echo "Stack ready. For Playwright E2E use:"
    echo "  export PLAYWRIGHT_BASE_URL=http://localhost:${FRONTEND_PORT}"
    echo "  export PLAYWRIGHT_API_URL=http://localhost:${BACKEND_PORT}"
    echo "Logs: ${RUN_DIR}/backend.log, ${RUN_DIR}/frontend.log"
    ;;

  stop)
    kill_group "${FRONTEND_PGID_FILE}" frontend
    kill_group "${BACKEND_PGID_FILE}" backend
    ;;

  status)
    for svc in backend frontend; do
      port_var=$([[ ${svc} == backend ]] && echo "${BACKEND_PORT}" || echo "${FRONTEND_PORT}")
      if curl -s -o /dev/null --max-time 2 "http://localhost:${port_var}/"; then
        echo "${svc}: up (port ${port_var})"
      else
        echo "${svc}: down (port ${port_var})"
      fi
    done
    ;;

  *)
    echo "Usage: $0 {start|stop|status}" >&2
    exit 1
    ;;
esac
