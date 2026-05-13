# Virtual Bingo Go Backend

This is the production backend foundation for the autonomous Virtual Bingo platform. It is intentionally small right now: typed environment config, health/readiness endpoints, a version endpoint, and placeholders for auth, game, and database packages.

## Local Setup

```bash
cd /Users/anish/Downloads/Work/BingoGame
docker compose up -d postgres
```

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
cp .env.example .env
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default.

## Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/version`

## Test And Format

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
gofmt -w ./cmd ./internal
go test ./...
```

## Environment Variables

| Variable | Default | Purpose |
|---|---:|---|
| `PORT` | `8080` | HTTP port for the Go API. |
| `APP_ENV` | `development` | Runtime environment label. |
| `DATABASE_URL` | empty | Postgres connection string for upcoming persistence work. |
| `CORS_ALLOWED_ORIGINS` | empty | Comma-separated browser origins allowed to call the API. |

## Current Deferrals

The scaffold deliberately does not implement Microsoft Entra, Microsoft Graph, Azure OpenAI, Azure Speech, rewards, voice profiles, or Azure deployment. Those are planned after the first local API and Postgres foundation is stable.
