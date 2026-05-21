#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_PORT="${FRONTEND_PORT:-3004}"
API_URL="${NEXT_PUBLIC_API_URL:-http://localhost:18081/api/v1}"

cd "$ROOT_DIR/apps/frontend-demo"

echo "Starting frontend on http://localhost:${FRONTEND_PORT}"
echo "Using API: ${API_URL}"

NEXT_PUBLIC_API_URL="$API_URL" npm run dev -- -p "$FRONTEND_PORT"
