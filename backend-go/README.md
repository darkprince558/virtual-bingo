# Virtual Bingo Go Backend

This is the production backend foundation for the autonomous Virtual Bingo platform. It currently has typed environment config, health/readiness endpoints, a version endpoint, Postgres migrations, local seed data, a store layer, Entra-ready auth seams, generated content prep/review/lock storage, locked call decks, caller asset records, mock delivery attempts, lobby/profile state, structured theme approval, and the Production V1 API flow for game runs, allowed players, player join/rejoin, connection heartbeat, persisted cards, game start, called words, marks, claims, winners, and summary state.

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
- `GET /api/v1/config`
- `GET /api/v1/me`
- `POST /api/v1/games`
- `GET /api/v1/games?status=&scope=host|player|admin`
- `GET /api/v1/games/{gameID}`
- `GET /api/v1/games/code/{code}`
- `PATCH /api/v1/games/{gameID}`
- `POST /api/v1/games/{gameID}/start`
- `POST /api/v1/games/{gameID}/pause`
- `POST /api/v1/games/{gameID}/resume`
- `POST /api/v1/games/{gameID}/finish`
- `POST /api/v1/games/{gameID}/cancel`
- `GET /api/v1/games/{gameID}/settings`
- `PATCH /api/v1/games/{gameID}/settings`
- `POST /api/v1/games/{gameID}/content/prepare`
- `GET /api/v1/games/{gameID}/content`
- `PATCH /api/v1/games/{gameID}/content`
- `POST /api/v1/games/{gameID}/content/lock`
- `POST /api/v1/games/{gameID}/caller-assets/generate`
- `POST /api/v1/games/{gameID}/deliveries/player-invites`
- `GET /api/v1/games/{gameID}/deliveries`
- `POST /api/v1/games/{gameID}/lobby/open`
- `POST /api/v1/games/{gameID}/theme`
- `POST /api/v1/games/{gameID}/auto-mark/run`
- `GET /api/v1/games/{gameID}/host-snapshot`
- `GET /api/v1/games/{gameID}/activity`
- `GET /api/v1/games/{gameID}/players/me/snapshot`
- `PATCH /api/v1/games/{gameID}/players/me/profile`
- `GET /api/v1/games/{gameID}/players/me/preferences`
- `PATCH /api/v1/games/{gameID}/players/me/preferences`
- `GET /api/v1/games/{gameID}/players/me/claim-readiness`
- `POST /api/v1/games/{gameID}/players/me/card`
- `GET /api/v1/games/{gameID}/players/me/card`
- `PATCH /api/v1/games/{gameID}/players/me/card/cells/{cellID}`
- `POST /api/v1/games/{gameID}/players/me/heartbeat`
- `GET /api/v1/games/{gameID}/players/{playerID}/snapshot`
- `POST /api/v1/games/{gameID}/players/{playerID}/heartbeat`
- `GET /api/v1/games/{gameID}/events`
- `POST /api/v1/games/{gameID}/calls`
- `GET /api/v1/games/{gameID}/calls`
- `POST /api/v1/games/{gameID}/allowed-players`
- `POST /api/v1/games/{gameID}/allowed-players/bulk`
- `GET /api/v1/games/{gameID}/allowed-players`
- `PATCH /api/v1/games/{gameID}/allowed-players/{allowedPlayerID}`
- `DELETE /api/v1/games/{gameID}/allowed-players/{allowedPlayerID}`
- `POST /api/v1/games/{gameID}/players`
- `POST /api/v1/games/{gameID}/players/{playerID}/card`
- `GET /api/v1/games/{gameID}/players/{playerID}/card`
- `PATCH /api/v1/games/{gameID}/players/{playerID}/card/cells/{cellID}`
- `POST /api/v1/games/{gameID}/claims`
- `GET /api/v1/games/{gameID}/claims`
- `GET /api/v1/games/{gameID}/claims/{claimID}`
- `GET /api/v1/games/{gameID}/summary`
- `POST /api/v1/deliveries/{deliveryID}/retry`
- `POST /api/v1/themes/generate`
- `GET /api/v1/themes/{themeID}`
- `PATCH /api/v1/themes/{themeID}`
- `POST /api/v1/themes/{themeID}/approve`
- `POST /api/v1/themes/{themeID}/reject`
- `GET /api/v1/theme-assets`
- `GET /api/v1/word-sets`
- `POST /api/v1/word-sets`
- `GET /api/v1/word-sets/{wordSetID}`
- `PATCH /api/v1/word-sets/{wordSetID}`
- `POST /api/v1/word-sets/{wordSetID}/words`
- `PATCH /api/v1/word-sets/{wordSetID}/words/{wordID}`
- `DELETE /api/v1/word-sets/{wordSetID}/words/{wordID}`

API responses are wrapped as `{ "data": ... }` for success and `{ "error": { "code": "...", "message": "..." } }` for errors. The API accepts `X-Request-ID` and returns it on each response.

## Local Management API Notes

- `GET /api/v1/config` is public so the frontend can discover local backend capabilities before auth. It reports the service/version, `appEnv`, `authMode`, heartbeat/reconnect timing, and capability flags. Current true flags include `sseEvents`, `devAuth` when in dev mode, `entraReadyAuth` when in Entra-ready mode, `gameSettings`, `autoMark`, `aiContent`, `aiCaller`, `themeGenerator`, and `automation`. Voice claims, Teams app APIs, and rewards are explicitly flagged false.
- `GET /api/v1/me` is authenticated. It uses the same principal flow as other protected endpoints and upserts the current user before returning `id`, `externalSubject`, `email`, `displayName`, `role`, and `authMode`.
- `GET /api/v1/games/code/{code}` looks up public game codes case-insensitively and returns the same game-run shape as `GET /api/v1/games/{gameID}`.
- `GET /api/v1/games?scope=host|player|admin&status=` is authenticated. Host scope returns games hosted by the current host/admin user, player scope returns games where the current email is joined or allowlisted, and admin scope is admin-only.
- `PATCH /api/v1/games/{gameID}` allows `name`, `code`, `wordSetId`, `scheduledStartAt`, and `winningPattern` while the game is still `draft`, `scheduled`, `invites_sent`, or `lobby_open`. Codes are normalized to uppercase and winning patterns use the existing bingo pattern normalization.
- `GET /api/v1/games/{gameID}/settings` and `PATCH /api/v1/games/{gameID}/settings` are host/admin-only. Settings are lazily created with defaults in `game_run_settings` and currently store marking mode, player override permission, claim-readiness visibility, and safe placeholder fields for future voice claim, caller, and theme behavior.
- `POST /api/v1/games/{gameID}/content/prepare` is a host/admin-only manual automation hook for local testing the T-60 content prep job. It calls the configured Python AI service through `POST /ai/v1/game-prep`, or uses the local disabled AI client when `AI_SERVICE_ENABLED=false`.
- `GET /api/v1/games/{gameID}/content` returns the generated review package: topic, summary, current editable words, review deadline, lock state, provider metadata, and any generation error.
- `PATCH /api/v1/games/{gameID}/content` lets a host/admin edit topic, summary, words, and caller style before lock. Words are trimmed, blank words are rejected, case-insensitive duplicates are removed, and at least 24 unique words are required.
- `POST /api/v1/games/{gameID}/content/lock` is a host/admin-only manual automation hook for local testing the T-30 lock job. It freezes the final topic/summary/words, creates an approved `ai_generated` word set from the locked words, creates a deterministic locked call deck from a stored seed/version, associates the game run with that word set, and blocks later content edits.
- `POST /api/v1/games/{gameID}/caller-assets/generate` is a host/admin-only post-lock hook. It sends the locked deck to Python `POST /ai/v1/caller-assets/bulk` when enabled, or to the local disabled client when `AI_SERVICE_ENABLED=false`, then stores one `caller_assets` row per deck item. Missing or failed assets keep fallback text such as `Next word is {word}.` so live gameplay never waits on Azure Speech.
- `POST /api/v1/games/{gameID}/calls` uses the locked deck order when a deck exists and links the committed called word back to the deck item. Older manual/local games without a locked deck keep the existing active-word fallback.
- `POST /api/v1/games/{gameID}/deliveries/player-invites` creates local/mock email delivery attempts from the allowed-player list. Attempts are recorded as sent with `/join?code={CODE}` links; login and allowlist checks remain authoritative. `GET /api/v1/games/{gameID}/deliveries` lists attempts and `POST /api/v1/deliveries/{deliveryID}/retry` resets a failed/mock attempt through the same local delivery boundary.
- `POST /api/v1/games/{gameID}/lobby/open` is the manual T-10 automation hook. It supports `draft`, `scheduled`, and `invites_sent` into `lobby_open`; live start still uses the existing start endpoint.
- `PATCH /api/v1/games/{gameID}/players/me/profile` lets the current player set a safe stored avatar profile: fixed icon ID, hex avatar color, and short label. Arbitrary image URLs are not accepted.
- Theme APIs are host/admin-only: `POST /api/v1/themes/generate`, `GET/PATCH /api/v1/themes/{themeID}`, `POST /api/v1/themes/{themeID}/approve`, `POST /api/v1/themes/{themeID}/reject`, `POST /api/v1/games/{gameID}/theme`, and `GET /api/v1/theme-assets`. AI themes are structured tokens only; Go rejects scripts, arbitrary CSS/URLs, and unapproved asset IDs. Applying a theme updates game settings with `themeMode=ai_generated`.
- `GET/PATCH /api/v1/games/{gameID}/players/me/preferences` lets an authenticated player read or set their own optional marking-mode preference. If `allowPlayerMarkingModeChoice=false`, the game setting wins; if true, the player preference can override marking mode.
- `POST /api/v1/games/{gameID}/auto-mark/run` is host/admin-only and idempotently backfills marks for already-called words for players whose effective mode is `auto_mark`.
- `GET /api/v1/games/{gameID}/players/me/claim-readiness` is a player-owned UX helper. It reads persisted card/called-word state and reports whether the player appears ready to submit a supported claim; claim submission remains the authoritative validation path.
- Bulk allowlist import accepts either a raw array of `{ "email", "displayName" }` rows or `{ "players": [...] }`. It is all-or-nothing: duplicate, blank, or conflicting rows fail the whole request without partial inserts.
- Manual word-set management still accepts only manual/seed sources through the public word-set APIs. Generated content lock creates approved `ai_generated` word sets internally so card assignment can use the existing deterministic path.
- Player `/me` aliases resolve the player by the authenticated principal email in the game and reuse the same snapshot/card/heartbeat/mark logic as the playerID endpoints.
- `GET /api/v1/games/{gameID}/activity` reads committed game outbox events for host/admin activity feeds.

## Auth Modes

Auth mode is selected with `AUTH_MODE`.

- `AUTH_MODE=dev` is the local default. It keeps tests and smoke flows easy by building the internal principal from development headers.
- `AUTH_MODE=entra-ready` switches the backend to the same `auth.Authenticator` handler boundary but expects a bearer token verified by an injectable token verifier. The current runtime verifier is intentionally unconfigured/offline, so real Microsoft network calls, JWKS fetching, and Azure credentials are not required yet.

Development auth is header-based:

```text
X-Dev-User-Email: host@example.local
X-Dev-User-Name: Local Development Host
X-Dev-User-Role: host
```

The Entra-ready seam now includes config placeholders for tenant ID, client ID/audience, issuer, and JWKS URL through `ENTRA_TENANT_ID`, `ENTRA_CLIENT_ID`, `ENTRA_AUDIENCE`, `ENTRA_ISSUER`, and `ENTRA_JWKS_URL`. The auth package has a token verifier interface, claims-to-principal mapping, and role mapping. Future Microsoft Entra JWT validation should plug into that verifier without changing handlers or the service principal shape.

## AI Service Config

The Go backend owns game truth. The Python AI service only generates draft content for review.

```text
AI_SERVICE_ENABLED=false
AI_SERVICE_BASE_URL=http://localhost:8000
AI_SERVICE_TIMEOUT_SECONDS=10
```

When `AI_SERVICE_ENABLED=false`, the backend uses a deterministic local disabled client that returns enough placeholder words, caller lines, and safe theme tokens to exercise prepare/review/lock/caller/theme flows without Python, Azure, or network calls. When enabled, the Go client calls `POST {AI_SERVICE_BASE_URL}/ai/v1/game-prep`, `POST {AI_SERVICE_BASE_URL}/ai/v1/caller-assets/bulk`, `POST {AI_SERVICE_BASE_URL}/ai/v1/caller-assets`, and `POST {AI_SERVICE_BASE_URL}/ai/v1/themes/generate`.

Real scheduled execution is still deferred. The `prepare`, `lock`, `caller-assets/generate`, `deliveries/player-invites`, and `lobby/open` endpoints are manual admin hooks around service methods that a future Azure Container Apps Job, Azure Function timer, Service Bus worker, or in-process worker can call.

Authorization behavior is explicit in the service layer:

- `admin` and `host` can create/manage games, add allowed players, call words, list host claims, and fetch host snapshots.
- `player` can fetch or heartbeat only their own player snapshot when the authenticated email matches the player record.
- Missing or invalid auth returns `{ "error": { "code": "unauthorized", "message": "authentication is required" } }`.
- Authenticated users without the needed role/scope return `{ "error": { "code": "forbidden", "message": "you do not have permission to perform this action" } }`.

## Realtime Backbone

The Production V1 realtime path is Postgres-first:

- Gameplay mutations write committed `game_event_outbox` rows in the same database transactions as the state changes where possible.
- Event sequences are ordered per game run and exposed as SSE event IDs.
- Auto-mark writes compact `card.auto_marked` outbox events with scanned player count, affected player count, called-word count, marked-cell count, and mode.
- `GET /api/v1/games/{gameID}/events` streams committed events with standard `net/http` Server-Sent Events.
- The SSE endpoint supports `Last-Event-ID` or `?lastEventId=` resume, sends heartbeat comments, and exits when the request context is cancelled.
- SSE is delivery only. Postgres snapshots and persisted gameplay tables remain the source of truth.

Snapshot endpoints are intended for reconnect and screen hydration:

- `GET /api/v1/games/{gameID}/host-snapshot` returns the game run, game settings, status, current word, winning pattern, player count, players, called words, claims with validation results, and winners. It requires a dev-auth `host` or `admin` role for now.
- `GET /api/v1/games/{gameID}/players/{playerID}/snapshot` returns the game run, effective marking mode, player-choice/claim-readiness flags, status, current word, winning pattern, player, assigned card with marks when present, called words, that player's claims, and winners. Dev auth currently allows host/admin access or a matching player email. A successful player snapshot marks that player `online` and refreshes `last_seen_at`.
- `POST /api/v1/games/{gameID}/players/{playerID}/heartbeat` marks the player `online` and refreshes `last_seen_at`. It requires host/admin access or a matching player email.

Connection state is persisted on `players.connection_state` and `players.last_seen_at`. New joins start `online`; explicit rejoins refresh `last_seen_at` and write a committed `player.reconnected` outbox event. Snapshot/heartbeat refreshes avoid noisy outbox rows while a player is already online, but write `player.reconnected` when the stored state was offline/disconnected. The current SSE endpoint is game-level, so it does not fake player disconnect identity on stream close. A future frontend should call the heartbeat endpoint while a player card is open and refetch snapshots after important SSE events.

The API also runs a configurable stale-player sweeper:

```text
PLAYER_CONNECTION_TIMEOUT_SECONDS=90
PLAYER_CONNECTION_SWEEP_INTERVAL_SECONDS=30
PLAYER_CONNECTION_SWEEP_BATCH_SIZE=100
```

When an `online` player in an active lobby/live/paused game has not checked in before the timeout, the backend marks them `disconnected` and writes a committed `player.disconnected` event. When that player later rejoins, heartbeats, or fetches their player snapshot, the backend returns them to `online` and includes a `reconnectNotice` payload with the words called after their previous `lastSeenAt`:

```json
{
  "reconnectNotice": {
    "lastSeenAt": "2026-05-15T15:00:00Z",
    "missedCalledWords": [{ "word": "Smoke Word 01", "sequence": 1 }]
  }
}
```

Redis, Service Bus fanout, and Gorilla/WebSocket are intentionally deferred. The current target is one Go API instance proving 50-player playability with simple Postgres polling and small event payloads; clients should refetch snapshots after important events. Redis or Azure fanout should only be added after load testing shows this local polling design is the bottleneck or multi-instance deployment needs cross-process delivery.

Still deferred for Production V1: frontend wiring, real Microsoft Entra login/JWKS validation, real Microsoft Graph/Teams delivery, real email delivery, Azure Speech execution/storage integration, rewards, voice profile consent flows, microphone claim recordings, lobby chat, Redis or Service Bus fanout, Gorilla/WebSocket, and Azure deployment.

## Game Settings And Marking Modes

Game-run settings are durable in Postgres and created lazily when read or patched. Defaults are conservative: `markingMode=manual`, `allowPlayerMarkingModeChoice=false`, `showClaimReadiness=true`, `voiceClaimMode=off`, `voiceClaimAutoplay=false`, `callerMode=off`, and `themeMode=default`.

Marking modes:

- `manual`: players mark/unmark cells themselves. Existing mark endpoints continue to work.
- `assist`: the backend does not mark cells automatically, but snapshots expose effective mode and `claim-readiness` can help a future UI highlight possible claims.
- `auto_mark`: when a word is called, matching unmarked cells are marked in the same transaction for players whose effective mode is `auto_mark`. Matching is case-insensitive and whitespace-trimmed. Auto-mark never submits a claim and never creates winners outside the existing claim flow.

Voice claim fields remain safe placeholders. Caller text/audio metadata and theme tokens are now persisted behind interfaces, but this slice still does not upload microphones, call Azure Speech from Go, render frontend themes, or build real Teams behavior.

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

Seed the local development game after migrations:

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
psql "postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable" -f internal/db/seeds/local_dev.sql
```

If `psql` is not installed locally, run the seed through Docker from the repo root:

```bash
cd /Users/anish/Downloads/Work/BingoGame
docker exec -i virtual-bingo-postgres psql -U bingo -d virtual_bingo < backend-go/internal/db/seeds/local_dev.sql
```

The local seed creates one host, one reusable word set, one game template/run, and a few allowed players.

## Local API Smoke Flow

Start the API:

```bash
cd /Users/anish/Downloads/Work/BingoGame/backend-go
set -a; source .env; set +a
go run ./cmd/api
```

Read local capabilities and current identity:

```bash
curl -sS http://localhost:8080/api/v1/config

curl -sS http://localhost:8080/api/v1/me \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Name: Local Development Host' \
  -H 'X-Dev-User-Role: host'
```

Create a game using the seeded word set:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Name: Local Development Host' \
  -H 'X-Dev-User-Role: host' \
  -d '{"name":"Local API Game","wordSetId":"00000000-0000-0000-0000-000000000201"}'
```

Add an allowed player:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/allowed-players \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host' \
  -d '{"email":"alex@example.local","displayName":"Alex Local"}'
```

Bulk add allowed players. This commits all rows or none:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/allowed-players/bulk \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host' \
  -d '[{"email":"alex@example.local","displayName":"Alex Local"},{"email":"sam@example.local","displayName":"Sam Local"}]'
```

List games by current-user scope, or look one up by code:

```bash
curl -sS 'http://localhost:8080/api/v1/games?scope=host&status=lobby_open' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS http://localhost:8080/api/v1/games/code/LOCAL-API-GAME
```

Update editable game metadata before the game goes live:

```bash
curl -sS -X PATCH http://localhost:8080/api/v1/games/<game-id> \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host' \
  -d '{"name":"Local API Game Updated","winningPattern":"four_corners"}'
```

Prepare, review, edit, and lock generated content locally with the disabled AI client:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/content/prepare \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS http://localhost:8080/api/v1/games/<game-id>/content \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS -X PATCH http://localhost:8080/api/v1/games/<game-id>/content \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host' \
  -d '{"topic":"Edited Local Topic","words":["Standup","Sprint Planning","Code Review","Deployment","Retrospective","Client Review","Documentation","Bug Bash","Architecture Review","Coffee Chat","Pull Request","Standards Review","Lunch and Learn","Release Notes","Security Review","Accessibility","Mentorship","Knowledge Transfer","Retest","Backlog Grooming","Pair Programming","Environment Setup","Ticket Refinement","Design Review"]}'

curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/content/lock \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'
```

Join the player:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/players \
  -H 'Content-Type: application/json' \
  -d '{"email":"alex@example.local","displayName":"Alex Local"}'
```

Assign and fetch a card:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/card
curl -sS http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/card
```

Player screens should prefer the authenticated `/me` aliases:

```bash
curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/players/me/card \
  -H 'X-Dev-User-Email: alex@example.local' \
  -H 'X-Dev-User-Role: player'

curl -sS http://localhost:8080/api/v1/games/<game-id>/players/me/snapshot \
  -H 'X-Dev-User-Email: alex@example.local' \
  -H 'X-Dev-User-Role: player'
```

Create and edit manual word sets:

```bash
curl -sS -X POST http://localhost:8080/api/v1/word-sets \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host' \
  -d '{"name":"Manual Local Words","status":"draft","source":"manual","words":[{"word":"Planning"},{"word":"Launch"}]}'

curl -sS -X POST http://localhost:8080/api/v1/word-sets/<word-set-id>/words \
  -H 'Content-Type: application/json' \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host' \
  -d '{"word":"Retrospective"}'
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

curl -sS http://localhost:8080/api/v1/games/<game-id>/activity \
  -H 'X-Dev-User-Email: host@example.local' \
  -H 'X-Dev-User-Role: host'

curl -sS http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/snapshot \
  -H 'X-Dev-User-Email: alex@example.local' \
  -H 'X-Dev-User-Role: player'

curl -sS -X POST http://localhost:8080/api/v1/games/<game-id>/players/<player-id>/heartbeat \
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
| `AI_SERVICE_ENABLED` | `false` | Enables HTTP calls to the Python AI service when true; false uses deterministic local disabled AI behavior. |
| `AI_SERVICE_BASE_URL` | `http://localhost:8000` | Base URL for Python AI endpoints. |
| `AI_SERVICE_TIMEOUT_SECONDS` | `10` | Timeout for Python AI service requests. |
| `AZURE_TENANT_ID` | empty | Placeholder for future Microsoft Entra integration. |
| `AZURE_CLIENT_ID` | empty | Placeholder for future Microsoft Entra app/client ID. |
| `AZURE_SERVICE_BUS_NAMESPACE` | empty | Placeholder for future Azure Service Bus event publishing. |
| `APPLICATIONINSIGHTS_CONNECTION_STRING` | empty | Placeholder for future Application Insights telemetry. |

## Current Deferrals

The backend now owns deterministic Production V1 game state plus the first Postgres-backed realtime delivery path, autonomous content/caller/theme storage, and local mock delivery. It still deliberately defers frontend wiring, Microsoft Entra production auth, real Microsoft Graph/Teams/email delivery, Azure Speech execution, rewards, Redis or Service Bus fanout, Gorilla/WebSocket, voice profiles/consent, microphone claim recordings, lobby chat, and Azure deployment. Those integrations should stay behind clean interfaces until the backend loop and frontend integration are stable.
