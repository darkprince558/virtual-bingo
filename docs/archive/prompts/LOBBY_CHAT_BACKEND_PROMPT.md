# Lobby Chat Backend Prompt

Status: ready to use after core gameplay, AI prep, deck/assets, invites, lobby, and theme backend work are stable.

```text
You are working in:

/Users/anish/Downloads/Work/BingoGame

Current branch:
codex/project-foundation-docs

Project direction:
Virtual Bingo is a full-scale internal workplace bingo platform. The Go backend owns game truth, player access, audit history, and realtime events. Lobby chat is a social/lobby feature, not a source of game state and not a blocker for gameplay.

Goal:
Implement backend-only lobby chat for game runs, using Postgres for persistence and the existing game event outbox/SSE path for realtime delivery.

Do not implement:
- frontend UI
- WebSockets
- Redis fanout
- AI moderation enforcement
- Teams chat integration
- rich media uploads
- direct messages
- chat after-game social features unless the schema makes it cheap and disabled by default

Read these first:
- README.md
- docs/README.md
- docs/architecture/AI_SERVICE_INTEGRATION_PLAN.md
- docs/architecture/FULL_SCALE_FEATURE_API_BACKLOG.md
- docs/architecture/V1_IMPLEMENTATION_PLAN.md
- backend-go/README.md
- backend-go/internal/domain/types.go
- backend-go/internal/db/store.go
- backend-go/internal/game/service.go
- backend-go/internal/app/routes.go
- backend-go/internal/app/game_handlers.go
- backend-go/internal/app/management_handlers.go
- backend-go/internal/app/api_test.go
- backend-go/internal/db/migrations/

Hard constraints:
- Backend Go work only.
- Keep standard library net/http.
- Keep handlers thin.
- Use Postgres as source of truth.
- Broadcast chat events through existing event outbox/SSE.
- Respect game access and auth.
- Do not let chat affect card state, called words, claims, winners, or game lifecycle.
- Plain go test ./... must pass without TEST_DATABASE_URL.
- DB integration tests should skip when TEST_DATABASE_URL is not set.
- Preserve response envelope:
  - success: { "data": ... }
  - error: { "error": { "code": "...", "message": "..." } }

Implementation tasks:

1. Chat schema and domain
- Add migrations and domain/store types for chat.
- Suggested table: `game_chat_messages`.
- Store:
  - id
  - game_run_id
  - player_id nullable
  - user_id nullable
  - sender_email
  - sender_display_name
  - body
  - status: visible, deleted, hidden
  - created_at
  - updated_at
  - deleted_at
  - deleted_by_user_id nullable
- Suggested table: `game_chat_reports`.
- Store:
  - id
  - message_id
  - reporter_user_id/player_id/email
  - reason
  - created_at
- Keep message bodies plain text only.

2. Access rules
- Host/admin can read all chat messages for a game.
- Allowlisted/joined players can read visible chat messages for their game.
- Players can create messages only as themselves.
- Host/admin can delete/hide any message.
- A player can delete their own message if product rules allow; if uncertain, allow host/admin only for delete.
- Disallow chat for finished/cancelled games unless the game settings explicitly allow it.
- Default intended use is lobby/live only.

3. Chat APIs
- Add routes:
  - GET /api/v1/games/{gameID}/chat/messages
  - POST /api/v1/games/{gameID}/chat/messages
  - DELETE /api/v1/games/{gameID}/chat/messages/{messageID}
  - POST /api/v1/games/{gameID}/chat/messages/{messageID}/report
- GET supports basic pagination:
  - `limit`
  - `before`
  - maybe `after`
- POST body:
  - `body`
- DELETE should soft-delete/hide the message.
- Report body:
  - `reason`
- Validate body length and reject blank messages.
- Suggested max message length: 500 chars.

4. Realtime events
- Write game outbox events:
  - `chat.message_created`
  - `chat.message_deleted`
  - `chat.message_reported`
- Payloads should be compact. Clients can refetch chat/messages after events.
- Do not introduce WebSockets in this pass.

5. Rate limiting / abuse controls
- Add simple local DB-backed or in-memory per-player message throttle if there is already a local pattern.
- If no existing rate limiter exists, implement conservative service-level validation that is testable:
  - max messages per player per rolling minute, or
  - minimum seconds between messages.
- Keep it small and documented. Do not add Redis.
- Add clear error code/message for rate limited chat.

6. Docs
- Update backend-go/README.md with chat endpoints and auth rules.
- Update docs/architecture/AI_SERVICE_INTEGRATION_PLAN.md lobby chat section if implementation differs.
- Update docs/architecture/FULL_SCALE_FEATURE_API_BACKLOG.md if chat is now part of the planned API surface.

Tests required:
- Unit/service tests for message validation and rate limiting.
- DB/API integration tests with TEST_DATABASE_URL for:
  - player can create/read chat in an accessible game
  - unauthorized/non-allowlisted user cannot read/post
  - host/admin can delete/hide message
  - deleted/hidden messages are not returned as visible
  - report creates a report row
  - outbox events are written for create/delete/report
  - chat is rejected in finished/cancelled games unless explicitly allowed
- Plain go test ./... must pass without TEST_DATABASE_URL.

Verification commands:
Run from repo root:

docker compose up -d postgres

Run from backend:

cd /Users/anish/Downloads/Work/BingoGame/backend-go
gofmt -w ./cmd ./internal
DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go run ./cmd/migrate up
TEST_DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go test ./...
go test ./...

Optional local API smoke:

PORT=18081 APP_ENV=test DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable CORS_ALLOWED_ORIGINS=http://localhost:3000 go run ./cmd/api

Smoke with curl:
- create game
- add allowed player
- join player
- open lobby/start if needed
- post chat as player
- list chat as player
- list chat as host
- delete/hide message as host
- report message
- verify SSE emits chat events or outbox rows exist

Commit discipline:
Make small commits if possible:
1. chat schema/store/domain
2. chat service/access/rate-limit
3. chat HTTP handlers/routes
4. outbox events/tests
5. docs

End with:
- files changed
- commands run
- tests passed
- API smoke summary
- remaining chat/frontend gaps only
```
