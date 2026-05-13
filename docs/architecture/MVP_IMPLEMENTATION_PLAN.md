# Virtual Bingo MVP Implementation Plan

Last updated: 2026-05-12

This plan turns the architecture docs into a practical two-week production MVP path for two students. The goal is not to build every autonomous feature immediately. The goal is to create a real backend/frontend foundation that can support the autonomous Azure platform later without pretending the whole enterprise system exists on day one.

## Two-Week MVP Scope

The MVP should prove one complete local game flow:

1. A host can create a simple game run.
2. Players can join the run through the web app.
3. The backend assigns or returns persisted bingo cards.
4. The host can start the game.
5. Called words are stored by the backend.
6. Players can mark cards and submit Bingo claims.
7. The backend validates claims deterministically.
8. The backend records the top 3 winners.
9. The frontend can show live-ish state through polling first, then SSE if time allows.
10. Local Postgres stores game data so a restart does not erase the session.

This MVP is local-first. Azure deployment is intentionally deferred until the core API, schema, and frontend integration are working.

## Backend And Frontend Split

### Backend Owns

- Typed config and local service startup.
- Postgres connection and migrations.
- Game run lifecycle.
- Player join/rejoin state.
- Bingo card generation and persistence.
- Word calling history.
- Mark state persistence.
- Server-side claim validation.
- Winner ordering and game summary data.
- API contracts for the frontend.

### Frontend Owns

- Host flow screens for creating and running a game.
- Player lobby and card screens.
- Claim button and claim result display.
- Winner and summary presentation.
- Calling backend APIs with simple development auth headers or local user selectors.
- Keeping the existing manager-demo styling useful without treating it as final production UI.

## Backend Milestone Order

1. Scaffold Go service and local Postgres.
2. Add database connection lifecycle and health/readiness checks that reflect DB availability.
3. Add migrations for users, game runs, players, word banks, cards, calls, claims, and winners.
4. Add development auth principal handling.
5. Add game run CRUD endpoints.
6. Add player join/rejoin endpoints.
7. Add card generation and assignment service.
8. Add word call endpoint and called-word history.
9. Add mark-card endpoint.
10. Add claim validation service and endpoints.
11. Add winner summary endpoints.
12. Add SSE endpoint for game events if polling becomes painful.

## Frontend Milestone Order

1. Point the demo app at the Go backend through environment-based API configuration.
2. Replace mock host game creation with backend-backed game runs.
3. Replace mock lobby/player data with backend join/rejoin APIs.
4. Replace mock cards with backend-assigned persisted cards.
5. Wire host call controls to backend called-word APIs.
6. Wire player marks and Bingo claims to backend APIs.
7. Show claim result and top 3 winners from backend state.
8. Add simple polling or SSE for live state updates.
9. Keep visual polish focused on clarity, not a redesign.

## First API Endpoints After Scaffold

Build these next, in this order:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Process health. Already scaffolded. |
| `GET` | `/readyz` | Dependency readiness. Next step: include Postgres ping. |
| `GET` | `/api/v1/version` | Service version and environment. Already scaffolded. |
| `POST` | `/api/v1/games` | Create a local MVP game run. |
| `GET` | `/api/v1/games/{gameID}` | Fetch game run state. |
| `POST` | `/api/v1/games/{gameID}/players` | Join or rejoin a player. |
| `GET` | `/api/v1/games/{gameID}/players/{playerID}/card` | Fetch assigned card. |
| `POST` | `/api/v1/games/{gameID}/start` | Start the game. |
| `POST` | `/api/v1/games/{gameID}/calls` | Record the next called word. |
| `POST` | `/api/v1/games/{gameID}/marks` | Toggle a player card mark. |
| `POST` | `/api/v1/games/{gameID}/claims` | Submit and validate a Bingo claim. |
| `GET` | `/api/v1/games/{gameID}/summary` | Return winners and final state. |

Use `net/http` until route complexity proves a small router is worth adding.

## Schema Phases

### Phase 1: Local MVP Tables

- `users`
- `game_runs`
- `players`
- `word_banks`
- `word_bank_words`
- `bingo_cards`
- `bingo_card_cells`
- `called_words`
- `claims`
- `winners`

### Phase 2: Autonomous Scheduling

- `game_templates`
- `template_audiences`
- `game_run_audience_snapshots`
- `invite_batches`
- `invite_deliveries`
- `automation_jobs`

### Phase 3: AI Content Review

- `prompt_libraries`
- `content_generation_jobs`
- `generated_word_sets`
- `content_approvals`
- `content_audit_events`

### Phase 4: Enterprise Integrations

- `entra_identities`
- `host_privilege_requests`
- `graph_delivery_connections`
- `voice_profiles`
- `voice_consents`
- `reward_providers`
- `reward_fulfillments`
- `security_audit_logs`

## Deferred Or Mocked On Purpose

These are not part of the first two-week MVP:

- Microsoft Entra ID production authentication.
- Microsoft Graph Outlook or Teams delivery.
- Azure OpenAI content generation.
- Azure Speech voice calling.
- Gift card or voucher fulfillment.
- Real voice cloning or employee voice profiles.
- Azure deployment and managed identities.
- Durable Azure Service Bus workflows.
- Redis fanout.
- Full admin console.

Use local development placeholders where needed. For example, development auth can identify the current user through headers or a local selector, generated word banks can be static, and winner notifications can be visible in the app instead of emailed.

## Immediate Next Task

Add Postgres connectivity to `backend-go`, update `/readyz` to ping the database when `DATABASE_URL` is set, and create the first migration for the local MVP schema skeleton.
