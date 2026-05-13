# Virtual Bingo Go Backend

This is the production backend foundation for the autonomous Virtual Bingo platform. It currently has typed environment config, health/readiness endpoints, a version endpoint, Postgres migrations, local seed data, a store layer, and the MVP API flow for game runs, allowed players, player join, persisted cards, game start, called words, marks, claims, winners, and summary state.

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
- `POST /api/v1/games/{gameID}/start`
- `POST /api/v1/games/{gameID}/pause`
- `POST /api/v1/games/{gameID}/resume`
- `POST /api/v1/games/{gameID}/finish`
- `POST /api/v1/games/{gameID}/cancel`
- `POST /api/v1/games/{gameID}/calls`
- `GET /api/v1/games/{gameID}/calls`
- `POST /api/v1/games/{gameID}/allowed-players`
- `GET /api/v1/games/{gameID}/allowed-players`
- `POST /api/v1/games/{gameID}/players`
- `POST /api/v1/games/{gameID}/players/{playerID}/card`
- `GET /api/v1/games/{gameID}/players/{playerID}/card`
- `PATCH /api/v1/games/{gameID}/players/{playerID}/card/cells/{cellID}`
- `POST /api/v1/games/{gameID}/claims`
- `GET /api/v1/games/{gameID}/claims`
- `GET /api/v1/games/{gameID}/summary`

API responses are wrapped as `{ "data": ... }` for success and `{ "error": { "code": "...", "message": "..." } }` for errors. The API accepts `X-Request-ID` and returns it on each response.

Development auth is header-based for now:

```text
X-Dev-User-Email: host@example.local
X-Dev-User-Name: Local Demo Host
X-Dev-User-Role: host
```

This is intentionally shaped so Microsoft Entra JWT validation can later produce the same internal principal without changing handlers.

## Game Rules Implemented

The backend currently supports these deterministic winning patterns:

- `single_line`: any row, column, or diagonal.
- `four_corners`: the four corner cells.
- `full_house`: every cell on the 5x5 card.

If `game_runs.winning_pattern` is set, submitted claims must use that exact pattern. If a game has no configured winning pattern, the MVP default is `single_line`; explicit `four_corners` or `full_house` claims are rejected until the host creates the game with that winning pattern.

Claim validation is pure Go logic in `internal/bingo`. Free spaces count automatically, and non-free cells count only when their word has already been called by the backend. Claim submission is transactional: claim persistence, validation result, winner placement, player-state updates, third-winner game finish, and audit rows commit or roll back together. Events are published only after the database transaction commits.

Winner placement is serialized per game run and remains limited to the top 3. Repeating the same valid player/pattern claim returns the existing winner placement instead of creating a duplicate winner row; the same player can still win different supported patterns in games that allow those patterns.

Lifecycle rules:

- `start`: `draft`, `scheduled`, `invites_sent`, or `lobby_open` -> `live`; joined/waiting players move to `playing`.
- `pause`: `live` -> `paused`.
- `resume`: `paused` -> `live`.
- `finish`: `live` or `paused` -> `finished`; `ended_at` is set if missing.
- `cancel`: `draft`, `scheduled`, `invites_sent`, `lobby_open`, `live`, or `paused` -> `cancelled`; `ended_at` is set if the game had started.
- Words can only be called while `live`.
- Claims are accepted for validation while `live` or `paused`, matching the architecture docs.
- When all active word-set words have been called, another call returns `409 conflict`; the game does not auto-finish from word exhaustion alone.

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

Start the game:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/start \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'
```

Pause and resume the game:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/pause \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/resume \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'
```

Call words. Repeat this command to advance the deterministic MVP caller through the active word set in `sort_order`:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/calls \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS http://localhost:8080/api/v1/games/<game-id>/calls
```

Mark or unmark a card cell:

```bash
curl -sS -X PATCH http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/card/cells/<cell-id> \
  -H 'Content-Type: application/json' \
  -d '{"marked":true}'
```

Submit a claim. The MVP validates `single_line` immediately on the backend and creates a top-3 winner row when valid:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/claims \
  -H 'Content-Type: application/json' \
  -d '{"playerId":"<player-id>","pattern":"single_line"}'
```

Fetch host claim state and game summary:

```bash
curl -sS http://localhost:8080/api/v1/games/<game-id>/claims \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS http://localhost:8080/api/v1/games/<game-id>/summary
```

Finish or cancel a game:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/finish \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS -X POST http://localhost:8080/api/v1/games/<other-game-id>/cancel \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'
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

The backend now owns the deterministic local MVP game state, but it still deliberately defers Microsoft Entra production auth, Microsoft Graph delivery, Teams automation, email delivery, AI caller behavior, Azure Speech, rewards, SSE/realtime fanout, voice profiles, Azure deployment, and frontend wiring. Those integrations should stay behind clean interfaces until the local game loop and frontend integration are stable.
