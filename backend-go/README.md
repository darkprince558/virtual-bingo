# Virtual Bingo Go Backend

This is the production backend foundation for the autonomous Virtual Bingo platform. It currently has typed environment config, health/readiness endpoints, a version endpoint, Postgres migrations, local seed data, and a small store layer for the first MVP game-run/card flow.

## Local Setup

```bash
cd /Users/anish/Downloads/Work/BingoGame
docker compose up -d postgres
```

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
cp .env.example .env
DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go run ./cmd/migrate up
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default.

## Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/version`

## Database

Start local Postgres from the repo root:

```bash
cd /Users/anish/Downloads/Work/BingoGame
docker compose up -d postgres
```

Run migrations from `backend-go`:

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go run ./cmd/migrate up
```

Roll migrations all the way down:

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go run ./cmd/migrate down
```

Seed the local demo game after migrations:

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
psql "postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable" -f internal/db/seeds/local_demo.sql
```

The local seed creates one host, one reusable word set, one game template/run, and a few allowed players.

## Test And Format

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
gofmt -w ./cmd ./internal
go test ./...
```

Run Postgres integration tests by pointing `TEST_DATABASE_URL` at the local Postgres server. The tests create and drop isolated test databases.

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
TEST_DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go test ./...
```

## Environment Variables

| Variable | Default | Purpose |
|---|---:|---|
| `PORT` | `8080` | HTTP port for the Go API. |
| `APP_ENV` | `development` | Runtime environment label. |
| `DATABASE_URL` | empty | Postgres connection string for API readiness and future persistence-backed endpoints. |
| `CORS_ALLOWED_ORIGINS` | empty | Comma-separated browser origins allowed to call the API. |
| `TEST_DATABASE_URL` | empty | Optional Postgres URL used by integration tests. |

## Current Deferrals

The scaffold deliberately does not implement gameplay endpoints, Microsoft Entra, Microsoft Graph, Azure OpenAI, Azure Speech, rewards, voice profiles, or Azure deployment. Those are planned after the first local API and Postgres foundation is stable.
