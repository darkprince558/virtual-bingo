# Production V1 Realtime Backbone Prompt

Status: completed by the Go backend realtime backbone pass. Keep this prompt as the source prompt/evidence for what was implemented and verified.

```text
You are working in:

/Users/anish/Downloads/Work/BingoGame

Current branch:
codex/project-foundation-docs

Project direction:
Virtual Bingo is no longer being treated as a small prototype. Build toward a full-scale Production V1 application. The target is a real internal workplace bingo platform that can support realtime playability for at least 50 players per game, then grow into autonomous weekly game operations, Microsoft 365 integration, AI-generated content, AI voice calling, audit history, rewards, and Azure deployment.

Important current decision:
Do NOT jump straight into AI voice, Teams, Graph, Redis, or frontend wiring. The next production step is the realtime gameplay backbone. AI voice depends on the realtime event pipeline anyway.

Read these first:
- README.md
- docs/README.md
- docs/architecture/V1_IMPLEMENTATION_PLAN.md
- docs/architecture/GO_BACKEND_PLAN.md
- docs/architecture/AUTONOMOUS_BACKEND_ARCHITECTURE.md
- backend-go/README.md
- backend-go/internal/game/service.go
- backend-go/internal/db/store.go
- backend-go/internal/domain/types.go
- backend-go/internal/app/routes.go
- backend-go/internal/app/game_handlers.go
- backend-go/internal/app/api_test.go
- backend-go/internal/bingo/validation.go
- backend-go/internal/db/migrations/

Current backend already supports:
- health/version/readiness
- game creation/fetch
- allowed players
- player join/rejoin
- card assignment/fetch
- start/pause/resume/finish/cancel lifecycle
- call words/list calls
- mark/unmark card cells
- submit/list claims
- pure bingo validation for single_line, four_corners, full_house
- allowed game pattern enforcement
- transactional claim/winner flow
- player state updates
- top-3 winners
- game summary
- Postgres migrations and integration tests that skip when TEST_DATABASE_URL is not set

Hard constraints:
- Backend Go work only for this pass unless explicitly told otherwise.
- Do not touch or rewrite the frontend.
- Do not implement Microsoft Entra yet.
- Do not implement Microsoft Graph, Teams, email, AI, voice, rewards, Azure deployment, or frontend wiring yet.
- Do not add Redis yet unless a load test proves it is necessary.
- Do not add Gorilla/WebSocket yet.
- Keep standard library net/http.
- Keep handlers thin.
- Keep deterministic game rules out of handlers.
- Keep authoritative gameplay data in Postgres.
- Keep response shape:
  - success: { "data": ... }
  - error: { "error": { "code": "...", "message": "..." } }
- Use existing dev auth and request ID patterns.
- Preserve Azure future seams behind interfaces only.
- Maintain go test ./... skipping DB integration tests when TEST_DATABASE_URL is not set.

Goal:
Implement the Production V1 realtime gameplay backbone for 50-player live play.

Build these backend pieces:

1. Host snapshot API
- Add GET /api/v1/games/{gameID}/host-snapshot
- Return one host-friendly payload with:
  - gameRun/status/currentWord/winningPattern
  - playerCount
  - players with useful state fields
  - calledWords
  - claims with validationResult
  - winners
  - enough fields for a future host screen without stitching many endpoints
- Require host/admin through dev auth for now.

2. Player snapshot API
- Add GET /api/v1/games/{gameID}/players/{playerID}/snapshot
- Return:
  - gameRun/status/currentWord/winningPattern
  - player
  - player card with marks
  - calledWords
  - player's claims
  - winners
  - enough fields for a future player screen to reconnect cleanly
- For now use dev auth patterns already in the app. Do not build Entra.

3. Event outbox
- Add a Postgres-backed event outbox table.
- Store committed gameplay events such as:
  - game.created
  - game.started
  - game.paused
  - game.resumed
  - game.finished
  - game.cancelled
  - player.joined
  - card.assigned
  - card.cell_marked
  - word.called
  - claim.submitted
  - claim.validated
  - winner.created
- Events should include:
  - id
  - game_run_id
  - type
  - entity_id
  - payload jsonb
  - sequence
  - created_at
- Event sequence must be ordered per game.
- Do not publish events before DB commit.
- Prefer writing outbox rows in the same transactions as the state changes where possible.
- Keep realtime publishing as an interface so Redis/Service Bus can be added later.

4. SSE endpoint
- Add GET /api/v1/games/{gameID}/events
- Use Server-Sent Events with standard net/http.
- No Gorilla/WebSocket.
- The endpoint should:
  - send committed outbox events in order
  - support Last-Event-ID if practical
  - send heartbeat comments to keep connections alive
  - close cleanly on request context cancellation
  - avoid busy database loops
- It is acceptable for V1 to use lightweight Postgres polling at first if documented and tested.
- Payloads should stay small. Clients can refetch snapshots after important events.

5. Service/store cleanup
- Keep handlers thin.
- Put snapshot assembly in service/store layers, not handlers.
- Keep event/outbox details behind focused store/service methods.
- Keep Postgres authoritative. SSE is delivery, not source of truth.

6. Load test
- Add a backend load test or command/script that simulates at least 50 SSE clients connected to one game.
- It should exercise:
  - 50 player connections
  - word calls
  - mark bursts
  - claim submissions
  - reconnect/snapshot fetch
- It should be runnable locally against the Go API.
- Do not require it in plain go test ./... unless it is fast and deterministic.
- Document how to run it.

7. Tests
Add/update tests. Minimum:
- Plain go test ./... passes without TEST_DATABASE_URL.
- DB/API integration tests with TEST_DATABASE_URL cover:
  - host snapshot
  - player snapshot
  - outbox rows are created for lifecycle/calls/claims/winners
  - SSE endpoint streams events in order
  - Last-Event-ID or equivalent resume behavior if implemented
  - heartbeat/cancellation does not hang tests
  - 50-player load helper can run locally or has a focused unit/integration smoke check

8. Docs
Update backend-go/README.md and docs/architecture/V1_IMPLEMENTATION_PLAN.md:
- snapshot endpoints
- SSE endpoint
- event outbox behavior
- 50-player load test command
- why Redis and Gorilla are deferred
- current deferrals remain:
  - Entra
  - Graph
  - Teams/email
  - AI caller
  - Azure Speech
  - rewards
  - Azure deployment
  - frontend wiring

Verification required:
Run from repo root:

docker compose up -d postgres

Run from backend:

cd /Users/anish/Downloads/Work/BingoGame/backend-go
gofmt -w ./cmd ./internal
DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go run ./cmd/migrate up
TEST_DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go test ./...
go test ./...

Then start API:

PORT=18081 APP_ENV=test DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable CORS_ALLOWED_ORIGINS=http://localhost:3000 go run ./cmd/api

Run curl smoke tests for:
- create game
- add allowed players
- join players
- assign cards
- start
- open SSE stream
- call words
- verify SSE receives word.called
- fetch host snapshot
- fetch player snapshot
- submit valid claim
- verify claim/winner event
- fetch summary
- pause/resume/finish/cancel lifecycle events

Commit discipline:
Make small commits:
1. event outbox schema/store methods
2. snapshot service/store methods
3. snapshot HTTP handlers/routes
4. SSE service/handler
5. realtime/load-test helper
6. expanded integration tests
7. README/docs updates

End with:
- files changed
- commits created
- commands run
- tests passed
- smoke test summary
- remaining Production V1 backend gaps only
```
