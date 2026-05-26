# Virtual Bingo Go Backend Plan

Last updated: 2026-05-12

Related planning docs:

- `../proposals/TECH_STACK_PROPOSAL_UPDATED.md`
- `AUTONOMOUS_BACKEND_ARCHITECTURE.md`
- `../archive/architecture/PRODUCTION_READINESS_RESEARCH.md`

## Direction

Build a separate Go backend service for the production game engine. Keep the current Next.js app as the frontend, and have it talk to Go over HTTP plus WebSocket/SSE for live updates.

This gives the bingo system a clean backend boundary for realtime game state, claim validation, audit logs, and future deployment scaling.

Important update:

- The backend is not only a live game API. The product direction now centers on autonomous weekly game operations: recurring game templates, scheduled game runs, Microsoft Teams/Outlook delivery, AI-generated content, AI voice calling, automatic winner validation, and reward fulfillment.
- See `AUTONOMOUS_BACKEND_ARCHITECTURE.md` for the higher-level architecture. This file remains the lower-level Go service/game-engine implementation plan.

## Backend Responsibilities

The Go service owns:

- Authentication verification and role authorization.
- Game/session creation and lifecycle.
- Player join/reconnect.
- Bingo card generation and persistence.
- Word bank management and word calling.
- Server-side bingo claim validation.
- Winner ordering and leaderboard state.
- Realtime event fanout to host/player screens.
- Audit logging and summary/export data.
- Health checks, metrics, and operational logs.

The Next.js frontend owns:

- UI rendering.
- Form interactions.
- Client-side optimistic/pending states.
- Microsoft sign-in UX if we keep Entra auth on the frontend.
- Calling backend APIs with an access token/session token.

## Recommended Stack

- Language: Go.
- HTTP router: `net/http` plus a small router such as `chi`.
- Database: Postgres.
- DB access: `pgx` with `sqlc` generated query methods.
- Migrations: `golang-migrate`.
- Realtime v1: native WebSocket endpoint in Go.
- Realtime scale-up: Redis pub/sub for multi-instance fanout.
- Config: environment variables loaded into a typed config struct.
- Auth v1: development bypass or signed local session token.
- Auth production: Microsoft Entra ID JWT verification and role mapping.
- Tests: Go unit tests for domain logic, integration tests for API/database behavior.

Avoid starting with a heavy ORM. The app is transactional and audit-heavy, so explicit SQL plus generated types should stay easier to reason about.

## Proposed Repo Layout

```text
backend/
  cmd/
    api/
      main.go
  internal/
    app/
      server.go
      routes.go
      middleware.go
    auth/
      principal.go
      entra.go
      dev.go
    bingo/
      card.go
      patterns.go
      validation.go
      wordbank.go
    config/
      config.go
    db/
      migrations/
      queries/
      sqlc/
      store.go
    realtime/
      hub.go
      client.go
      events.go
    sessions/
      service.go
      handlers.go
      snapshots.go
    claims/
      service.go
      handlers.go
    audit/
      service.go
    exports/
      csv.go
  test/
    fixtures/
  go.mod
  sqlc.yaml
  Dockerfile
```

## Core Domain Model

### Users

Represents an authenticated person or a temporary guest player.

Fields:

- `id`
- `tenant_id`
- `external_subject`
- `display_name`
- `email`
- `role`
- `created_at`
- `updated_at`

Roles:

- `admin`
- `host`
- `player`
- `viewer`

### Tenants

Supports a future multi-company setup without forcing it into the UI right away.

Fields:

- `id`
- `name`
- `entra_tenant_id`
- `created_at`

For v1, seed one tenant.

### Game Sessions

Represents one bingo event.

Fields:

- `id`
- `tenant_id`
- `code`
- `name`
- `status`: `waiting`, `starting_soon`, `live`, `paused`, `finished`, `cancelled`
- `host_user_id`
- `word_bank_id`
- `winning_patterns`
- `current_called_word_id`
- `started_at`
- `ended_at`
- `created_at`
- `updated_at`

Constraints:

- Public `code` should be unique for active sessions.
- Host mutations should check `updated_at` or a numeric `version` to avoid stale writes.

### Players

Represents a player inside a session.

Fields:

- `id`
- `session_id`
- `user_id`
- `display_name`
- `connection_state`
- `state`: `joined`, `waiting`, `playing`, `claimed_bingo`, `confirmed_winner`, `rejected_claim`, `disconnected`
- `joined_at`
- `last_seen_at`

Constraints:

- One active player record per `session_id + user_id`.

### Word Banks

Reusable collections of bingo words.

Tables:

- `word_banks`
- `word_bank_words`

Fields for words:

- `id`
- `word_bank_id`
- `word`
- `sort_order`
- `is_active`

### Bingo Cards

Persist the exact card assigned to each player.

Tables:

- `bingo_cards`
- `bingo_card_cells`

Card fields:

- `id`
- `session_id`
- `player_id`
- `seed`
- `created_at`

Cell fields:

- `id`
- `card_id`
- `row`
- `col`
- `word`
- `is_free_space`
- `marked_at`

Important rule:

- The server stores card assignment and marks. The client can request a mark toggle, but the server remains source of truth.

### Called Words

Immutable word-call history.

Fields:

- `id`
- `session_id`
- `word`
- `called_by_user_id`
- `sequence`
- `called_at`

Constraints:

- Unique `session_id + sequence`.
- Unique `session_id + word`.

### Claims

Represents a player bingo claim.

Fields:

- `id`
- `session_id`
- `player_id`
- `pattern`
- `status`: `pending`, `valid`, `invalid`, `confirmed`, `rejected`
- `validation_result`
- `claimed_at`
- `reviewed_by_user_id`
- `reviewed_at`

Validation result should include:

- matched cells
- missing cells
- called words considered
- reason for invalid result

### Winners

Final confirmed winners.

Fields:

- `id`
- `session_id`
- `player_id`
- `claim_id`
- `placement`
- `pattern`
- `confirmed_at`

Constraints:

- Unique `session_id + placement`.
- Unique `session_id + player_id` if each player can win only once.

### Audit Events

Append-only event stream for traceability.

Fields:

- `id`
- `tenant_id`
- `session_id`
- `actor_user_id`
- `event_type`
- `entity_type`
- `entity_id`
- `payload_json`
- `created_at`

Examples:

- `session.created`
- `session.started`
- `player.joined`
- `word.called`
- `claim.submitted`
- `claim.validated`
- `claim.confirmed`
- `claim.rejected`
- `session.finished`
- `summary.exported`

## REST API Shape

Use `/api/v1` for all production backend routes.

### Health

- `GET /healthz`
- `GET /readyz`

### Auth / Identity

- `GET /api/v1/me`
- `POST /api/v1/dev-login` for local development only

### Sessions

- `POST /api/v1/sessions`
- `GET /api/v1/sessions/{sessionId}`
- `GET /api/v1/sessions/code/{code}`
- `POST /api/v1/sessions/{sessionId}/start`
- `POST /api/v1/sessions/{sessionId}/pause`
- `POST /api/v1/sessions/{sessionId}/resume`
- `POST /api/v1/sessions/{sessionId}/end`
- `GET /api/v1/sessions/{sessionId}/snapshot`
- `GET /api/v1/sessions/{sessionId}/summary`

### Players

- `POST /api/v1/sessions/{sessionId}/players`
- `GET /api/v1/sessions/{sessionId}/players`
- `GET /api/v1/sessions/{sessionId}/players/me`
- `POST /api/v1/sessions/{sessionId}/players/me/heartbeat`

### Cards

- `GET /api/v1/sessions/{sessionId}/players/me/card`
- `POST /api/v1/sessions/{sessionId}/players/me/card/cells/{cellId}/toggle`

### Word Calling

- `POST /api/v1/sessions/{sessionId}/calls/next`
- `GET /api/v1/sessions/{sessionId}/calls`

### Claims

- `POST /api/v1/sessions/{sessionId}/claims`
- `GET /api/v1/sessions/{sessionId}/claims`
- `POST /api/v1/sessions/{sessionId}/claims/{claimId}/confirm`
- `POST /api/v1/sessions/{sessionId}/claims/{claimId}/reject`

### Exports

- `GET /api/v1/sessions/{sessionId}/exports/audit.csv`
- `GET /api/v1/sessions/{sessionId}/exports/summary.csv`

## Snapshot Contract

The frontend should prefer one snapshot endpoint per screen to avoid stitching many requests together.

### Host Snapshot

```json
{
  "session": {
    "id": "uuid",
    "code": "BINGO-2026",
    "status": "live",
    "currentWord": "Action Item"
  },
  "players": [],
  "calledWords": [],
  "claims": [],
  "leaderboard": [],
  "activityEvents": []
}
```

### Player Snapshot

```json
{
  "session": {
    "id": "uuid",
    "code": "BINGO-2026",
    "status": "live",
    "currentWord": "Action Item"
  },
  "player": {},
  "card": {
    "id": "uuid",
    "cells": []
  },
  "calledWords": [],
  "leaderboard": [],
  "claimState": null
}
```

## Realtime Event Plan

Endpoint:

- `GET /api/v1/sessions/{sessionId}/events`

Recommended first version:

- Use WebSockets if we want bidirectional messages.
- Use Server-Sent Events if the client only receives events and sends mutations through REST.

For this app, SSE plus REST is enough for v1 and simpler to operate. WebSockets can come later if we need richer presence behavior.

Events:

- `session.updated`
- `player.joined`
- `player.left`
- `player.connection_changed`
- `word.called`
- `card.marked`
- `claim.submitted`
- `claim.validated`
- `claim.confirmed`
- `claim.rejected`
- `leaderboard.updated`
- `audit.created`
- `summary.ready`

Event envelope:

```json
{
  "id": "event-id",
  "type": "word.called",
  "sessionId": "uuid",
  "sequence": 12,
  "occurredAt": "2026-05-12T18:00:00Z",
  "payload": {}
}
```

The frontend should refetch the appropriate snapshot after important events. That keeps the realtime payloads small and avoids complicated client reconciliation.

## Claim Validation Rules

Server validation must not trust the client.

Inputs:

- persisted player card
- persisted marked cells
- persisted called words
- allowed session patterns
- submitted pattern

Required checks:

- Claim belongs to the session.
- Player belongs to the session.
- Game is live or paused, not finished.
- Player has an assigned card.
- Submitted pattern is allowed for this session.
- Pattern exists on the marked card.
- Every non-free-space word in the pattern has been called.

Initial patterns:

- Any row
- Any column
- Diagonal
- Four corners
- Full house

Recommended implementation:

- Put pure validation logic in `internal/bingo`.
- Unit test it heavily before wiring HTTP.
- Keep validation output structured so the host can see why a claim passed or failed.

## Transaction Boundaries

Use DB transactions for:

- Create session with word bank selection and audit event.
- Join session with player record, generated card, and audit event.
- Start/pause/resume/end session with audit event.
- Call next word with called-word row, session current word update, audit event, and realtime broadcast enqueue.
- Submit claim with validation result and audit event.
- Confirm winner with claim update, winner placement, player state update, audit event, and summary update.

Do not broadcast realtime updates until the database transaction commits.

## Auth Plan

### Local Development

Start with a dev auth middleware:

- `X-Dev-User-Id`
- `X-Dev-Role`
- `X-Dev-Display-Name`

This lets us build backend flows before the Entra setup is ready.

### Production

Use Microsoft Entra ID access tokens:

- Validate JWT issuer, audience, expiry, signature, and tenant.
- Map groups/app roles to app roles.
- Store or upsert user profile on first request.
- Require host/admin role for host mutations.
- Allow players to read only their own card.

Open decision:

- Guests by game code may be useful for local testing, but corporate production probably wants Entra-only or explicit guest-mode controls.

## Implementation Milestones

### Milestone 1: Backend Skeleton

Deliverables:

- `backend/` Go module.
- HTTP server with config, logger, health endpoints.
- Dockerfile.
- Local `.env.example`.
- Basic request logging and panic recovery.
- Dev auth middleware.

Done when:

- `go test ./...` passes.
- `GET /healthz` works locally.

### Milestone 2: Database Foundation

Deliverables:

- Postgres docker compose for local development.
- Migrations for tenants, users, sessions, players, word banks, cards, calls, claims, winners, audit events.
- `sqlc` query setup.
- Seed data matching the current local bingo flow.

Done when:

- Migrations run from empty DB.
- Seed command creates one playable session.
- Tests can reset and seed the DB.

### Milestone 3: Session And Player Flow

Deliverables:

- Create session.
- Join session by code.
- Generate/persist player card.
- Host/player snapshot endpoints.
- Frontend can use backend lobby/play/host snapshots directly.

Done when:

- Host can create a session.
- Player can join and refresh without losing card.

### Milestone 4: Word Calling And Audit

Deliverables:

- Start/pause/resume/end APIs.
- Call-next-word API.
- Immutable called-word history.
- Audit events.
- Summary endpoint reads persisted totals.

Done when:

- Host actions survive refresh.
- Called words are persisted and ordered.

### Milestone 5: Claims And Winners

Deliverables:

- Mark/toggle cells server-side.
- Submit claim.
- Server-side validation.
- Confirm/reject claim.
- Winner placement.
- Leaderboard calculation.

Done when:

- Player cannot win with an invalid card/pattern.
- Host sees pending/valid/invalid claim states.
- Summary shows real winners.

### Milestone 6: Realtime

Deliverables:

- SSE or WebSocket session event stream.
- Broadcast events after committed mutations.
- Frontend refetches snapshots on relevant events.
- Connection/reconnecting UI state backed by real connection status.

Done when:

- Host calling a word updates player screens without manual refresh.
- Player claim appears on host screen without manual refresh.

### Milestone 7: Production Hardening

Deliverables:

- Entra JWT validation.
- Role authorization.
- Rate limits.
- CORS policy.
- Security headers at frontend/proxy layer.
- Audit/summary CSV export.
- Metrics/logging.
- CI for Go tests, frontend lint/build, and E2E smoke.

Done when:

- App can be deployed to a production-like environment with a clean runbook.

## First Files To Create

Start here:

```text
backend/go.mod
backend/cmd/api/main.go
backend/internal/config/config.go
backend/internal/app/server.go
backend/internal/app/routes.go
backend/internal/app/middleware.go
backend/internal/auth/dev.go
backend/internal/bingo/patterns.go
backend/internal/bingo/validation.go
backend/internal/realtime/events.go
backend/internal/db/migrations/000001_initial_schema.up.sql
backend/internal/db/migrations/000001_initial_schema.down.sql
backend/sqlc.yaml
```

## Early Decisions To Make

- SSE-first or WebSocket-first?
- Entra-only players or guest code players for v1?
- One tenant only or multi-tenant schema from day one?
- Host-created custom word banks in v1 or seeded word banks only?
- Winning pattern defaults: line only, or line plus four corners/full house?
- Target scale for first production event: 50, 100, 250, or 500 players?

## Recommended Decisions For V1

- Use SSE plus REST first.
- Use Postgres from day one.
- Include tenant IDs in schema from day one, but seed one tenant.
- Use dev auth first, then Entra before production.
- Ship seeded word banks first, then add custom word-bank UI later.
- Support row, column, diagonal, four corners, and full house in backend logic, even if UI initially exposes fewer options.
- Target 100 concurrent players first, then load test upward.

## External References

- Go database docs: https://go.dev/doc/database/
- Go `database/sql` package: https://pkg.go.dev/database/sql
- sqlc docs: https://docs.sqlc.dev/
- golang-migrate: https://github.com/golang-migrate/migrate
- Microsoft MSAL / Entra auth docs: https://learn.microsoft.com/en-us/entra/msal/
