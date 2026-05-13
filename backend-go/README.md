# Virtual Bingo Go Backend

This is the production backend foundation for the autonomous Virtual Bingo platform. It currently has typed environment config, health/readiness endpoints, a version endpoint, Postgres migrations, local seed data, a store layer, and the first MVP API flow for game runs, allowed players, player join, and persisted cards.

## Local Setup

```bash
cd /Users/anish/Downloads/Work/BingoGame
docker compose up -d postgres
```

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
cp .env.example .env
set -a; source .env; set +a
DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go run ./cmd/migrate up
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default.

## Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/version`
- `POST /api/v1/games`
- `GET /api/v1/games/{gameID}`
- `POST /api/v1/games/{gameID}/allowed-players`
- `GET /api/v1/games/{gameID}/allowed-players`
- `POST /api/v1/games/{gameID}/players`
- `POST /api/v1/games/{gameID}/players/{playerID}/card`
- `GET /api/v1/games/{gameID}/players/{playerID}/card`

API responses are wrapped as `{ "data": ... }` for success and `{ "error": { "code": "...", "message": "..." } }` for errors. The API accepts `X-Request-ID` and returns it on each response.

Development auth is header-based for now:

```text
X-Dev-User-Email: host@example.local
X-Dev-User-Name: Local Demo Host
X-Dev-User-Role: host
```

This is intentionally shaped so Microsoft Entra JWT validation can later produce the same internal principal without changing handlers.

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

If `psql` is not installed locally, run the seed through Docker from the repo root:

```bash
cd /Users/anish/Downloads/Work/BingoGame
docker exec -i virtual-bingo-postgres psql -U bingo -d virtual_bingo < backend-go/internal/db/seeds/local_demo.sql
```

The local seed creates one host, one reusable word set, one game template/run, and a few allowed players.

## Local API Smoke Flow

Start the API:

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
set -a; source .env; set +a
go run ./cmd/api
```

Create a game using the seeded word set:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Name: Local Demo Host' \
  -H 'X-Dev-User-Role: host' \
  -d '{"name":"Local API Game","wordSetId":"00000000-0000-0000-0000-000000000201"}'
```

Add an allowed player:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/allowed-players \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host' \
  -d '{"email":"alex@example.local","displayName":"Alex Demo"}'
```

Join the player:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/players \
  -H 'Content-Type: application/json' \
  -d '{"email":"alex@example.local","displayName":"Alex Demo"}'
```

Assign and fetch a card:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/card
curl -sS http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/card
```

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
| `AZURE_TENANT_ID` | empty | Placeholder for future Microsoft Entra integration. |
| `AZURE_CLIENT_ID` | empty | Placeholder for future Microsoft Entra app/client ID. |
| `AZURE_SERVICE_BUS_NAMESPACE` | empty | Placeholder for future Azure Service Bus event publishing. |
| `APPLICATIONINSIGHTS_CONNECTION_STRING` | empty | Placeholder for future Application Insights telemetry. |

## Current Deferrals

The scaffold deliberately does not implement gameplay endpoints, Microsoft Entra, Microsoft Graph, Azure OpenAI, Azure Speech, rewards, voice profiles, or Azure deployment. Those are planned after the first local API and Postgres foundation is stable.
