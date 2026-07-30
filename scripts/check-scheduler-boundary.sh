#!/usr/bin/env bash
# Assert backend/scheduler stays a pure domain package:
# no net/http, no pgx, no imports of backend/api.
#
# Usage (from repo root or backend/):
#   ./scripts/check-scheduler-boundary.sh
#   (or) cd backend && ../scripts/check-scheduler-boundary.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKEND_DIR="${REPO_ROOT}/backend"

if [[ ! -d "${BACKEND_DIR}/scheduler" ]]; then
  echo "Error: backend/scheduler not found at ${BACKEND_DIR}/scheduler" >&2
  exit 1
fi

cd "${BACKEND_DIR}"

# go list -f prints each import path of the scheduler package (and its tests).
IMPORTS="$(go list -f '{{join .Imports "\n"}}{{"\n"}}{{join .TestImports "\n"}}{{"\n"}}{{join .XTestImports "\n"}}' ./scheduler/...)"

FORBIDDEN_PATTERNS=(
  '^net/http$'
  'pgx'
  '(^|/)backend/api(/|$)'
)

FAILED=0
while IFS= read -r imp; do
  [[ -z "${imp}" ]] && continue
  for pat in "${FORBIDDEN_PATTERNS[@]}"; do
    if printf '%s' "${imp}" | grep -Eq "${pat}"; then
      echo "FORBIDDEN import in scheduler: ${imp}" >&2
      FAILED=1
    fi
  done
done <<< "${IMPORTS}"

# Also catch blank/side-effect imports that go list may still surface via Deps,
# and direct source matches for clarity in CI logs.
if grep -RIn --include='*.go' -E '^\s*(_\s+)?"net/http"|^\s*(_\s+)?"[^"]*pgx[^"]*"|^\s*(_\s+)?"[^"]*/backend/api("|/)' scheduler/ >&2; then
  echo "FORBIDDEN import pattern found in scheduler source (see above)." >&2
  FAILED=1
fi

if [[ "${FAILED}" -ne 0 ]]; then
  echo "" >&2
  echo "Architecture boundary violated: backend/scheduler must not import" >&2
  echo "  net/http, pgx (jackc/pgx), or packages under backend/api." >&2
  echo "Keep the scheduler pure so it stays unit-testable in isolation." >&2
  exit 1
fi

echo "OK: backend/scheduler import boundary holds."
