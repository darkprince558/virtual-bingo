#!/usr/bin/env bash
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:18081}"
FRONTEND_BASE="${FRONTEND_BASE:-http://localhost:3004}"

echo "Checking Go API..."
curl -fsS "$API_BASE/healthz" >/dev/null
curl -fsS "$API_BASE/readyz" >/dev/null
curl -fsS "$API_BASE/api/v1/games/code/LOCAL-DEMO" >/dev/null

echo "Checking frontend demo route..."
curl -fsS "$FRONTEND_BASE/demo" >/dev/null

echo "Demo smoke passed:"
echo "- API: $API_BASE"
echo "- Frontend: $FRONTEND_BASE/demo"
