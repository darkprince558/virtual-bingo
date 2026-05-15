# Virtual Bingo Go Backend

This is the production backend foundation for the autonomous Virtual Bingo platform. It currently has typed environment config, health/readiness endpoints, a version endpoint, Postgres migrations, local seed data, a store layer, and the Production V1 API flow for game runs, allowed players, player join, persisted cards, game start, called words, marks, claims, winners, and summary state.

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
- `GET /api/v1/games/{gameID}/host-snapshot`
- `GET /api/v1/games/{gameID}/players/{playerID}/snapshot`
- `GET /api/v1/games/{gameID}/events`
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

## Realtime Backbone

The Production V1 realtime path is Postgres-first:

- Gameplay mutations write committed `game_event_outbox` rows in the same database transactions as the state changes where possible.
- Event sequences are ordered per game run and exposed as SSE event IDs.
- `GET /api/v1/games/{gameID}/events` streams committed events with standard `net/http` Server-Sent Events.
- The SSE endpoint supports `Last-Event-ID` or `?lastEventId=` resume, sends heartbeat comments, and exits when the request context is cancelled.
- SSE is delivery only. Postgres snapshots and persisted gameplay tables remain the source of truth.

Snapshot endpoints are intended for reconnect and screen hydration:

- `GET /api/v1/games/{gameID}/host-snapshot` returns the game run, status, current word, winning pattern, player count, players, called words, claims with validation results, and winners. It requires a dev-auth `host` or `admin` role for now.
- `GET /api/v1/games/{gameID}/players/{playerID}/snapshot` returns the game run, status, current word, winning pattern, player, assigned card with marks when present, called words, that player's claims, and winners. Dev auth currently allows host/admin access or a matching player email.

Redis, Service Bus fanout, and Gorilla/WebSocket are intentionally deferred. The current target is one Go API instance proving 50-player playability with simple Postgres polling and small event payloads; clients should refetch snapshots after important events. Redis or Azure fanout should only be added after load testing shows this local polling design is the bottleneck or multi-instance deployment needs cross-process delivery.

## Game Rules Implemented

The backend currently supports these deterministic winning patterns:

- `single_line`: any row, column, or diagonal.
- `four_corners`: the four corner cells.
- `full_house`: every cell on the 5x5 card.

If `game_runs.winning_pattern` is set, submitted claims must use that exact pattern. If a game has no configured winning pattern, the Production V1 default is `single_line`; explicit `four_corners` or `full_house` claims are rejected until the host creates the game with that winning pattern.

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

Call words. Repeat this command to advance the deterministic Production V1 caller through the active word set in `sort_order`:

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

Submit a claim. The Production V1 backend validates the claim immediately and creates a top-3 winner row when valid:

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

curl -sS http://localhost:8080/api/v1/games/<game-id>/host-snapshot \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/snapshot \
  -H 'X-Dev-User-Email: alex@example.local' \
  -H 'X-Dev-User-Role: player'

curl -sS http://localhost:8080/api/v1/games/<game-id>/summary
```

Open an SSE stream in another terminal. `Last-Event-ID` can be set by reconnecting clients to resume after the last processed sequence.

```bash
curl -N http://localhost:8080/api/v1/games/<game-id>/events \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'
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

Run the local 50-player realtime helper against a running API. It creates a game, joins players, opens 50 SSE streams, sends mark bursts, calls words, submits claims, and fetches reconnect snapshots.

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
go run ./cmd/realtime-loadtest \
  -base-url http://localhost:8080 \
  -word-set-id 00000000-0000-0000-0000-000000000201 \
  -players 50 \
  -word-calls 12
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

The backend now owns deterministic Production V1 game state plus the first Postgres-backed realtime delivery path. It still deliberately defers Microsoft Entra production auth, Microsoft Graph delivery, Teams automation, email delivery, AI caller behavior, Azure Speech, rewards, Redis or Service Bus fanout, Gorilla/WebSocket, voice profiles, Azure deployment, and frontend wiring. Those integrations should stay behind clean interfaces until the realtime game loop and frontend integration are stable.
