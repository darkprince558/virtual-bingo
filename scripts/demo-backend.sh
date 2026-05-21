#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
POSTGRES_PORT="${POSTGRES_PORT:-55432}"
DATABASE_URL="${DATABASE_URL:-postgres://bingo:bingo@localhost:${POSTGRES_PORT}/virtual_bingo?sslmode=disable}"
API_PORT="${API_PORT:-18081}"
FRONTEND_ORIGIN="${FRONTEND_ORIGIN:-http://localhost:3004}"

cd "$ROOT_DIR"

POSTGRES_PORT="$POSTGRES_PORT" docker compose up -d postgres

cd "$ROOT_DIR/backend-go"
DATABASE_URL="$DATABASE_URL" go run ./cmd/migrate up

docker exec -i virtual-bingo-postgres psql -U bingo -d virtual_bingo < "$ROOT_DIR/backend-go/internal/db/seeds/local_demo.sql"

echo "Demo seed loaded: LOCAL-DEMO"
echo "Starting Go API on http://localhost:${API_PORT}"

PORT="$API_PORT" \
APP_ENV=development \
AUTH_MODE=dev \
DATABASE_URL="$DATABASE_URL" \
CORS_ALLOWED_ORIGINS="$FRONTEND_ORIGIN,http://localhost:3000,http://localhost:3003" \
AI_SERVICE_ENABLED="${AI_SERVICE_ENABLED:-false}" \
AI_SERVICE_BASE_URL="${AI_SERVICE_BASE_URL:-http://localhost:5001}" \
AI_SERVICE_TIMEOUT_SECONDS="${AI_SERVICE_TIMEOUT_SECONDS:-10}" \
go run ./cmd/api
