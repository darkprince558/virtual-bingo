# Virtual Bingo Production V1 Implementation Plan

Last updated: 2026-05-15

This plan turns the architecture docs into a production V1 path for Virtual Bingo. The project is no longer being treated as a throwaway prototype. The goal is a full-scale ready application foundation: correct game state, realtime playability for at least 50 players per game, production authentication seams, autonomous operations, AI voice/content seams, auditability, and Azure-ready deployment discipline.

For the larger CGI-only Microsoft 365, Azure, AI caller, Redis/WebSocket, weekly automation, and rewards roadmap beyond this V1 game foundation, see `FULL_SCALE_DEPLOYMENT_ROADMAP.md`.

## Production V1 Scope

Production V1 should support one real internal game flow end to end:

1. A host can create a simple game run.
2. Players can join the run through the web app.
3. The backend assigns or returns persisted bingo cards.
4. The host can start the game.
5. Called words are stored by the backend.
6. Players can mark cards and submit Bingo claims.
7. The backend validates claims deterministically.
8. The backend records the top 3 winners.
9. The frontend can show realtime state through SSE plus snapshot refetching.
10. Local Postgres stores game data so a restart does not erase the session.
11. The game engine is load-tested for at least 50 connected players per game.
12. Production integrations stay behind interfaces until each one is implemented for real.

Production V1 is still built locally first, but local work should be shaped for deployment. Azure deployment can wait until core API, schema, realtime, auth, and frontend integration are credible, but code should not assume a single laptop process forever.

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

1. Done: scaffold Go service and local Postgres.
2. Done: add database connection lifecycle and health/readiness checks that reflect DB availability.
3. Done: add migrations for users, game runs, players, word sets, cards, calls, claims, winners, audit events, and game event outbox.
4. Done: add development auth principal handling.
5. Done: add game run CRUD endpoints.
6. Done: add player join/rejoin endpoints.
7. Done: add card generation and assignment service.
8. Done: add word call endpoint and called-word history.
9. Done: add mark-card endpoint.
10. Done: add claim validation service and endpoints.
11. Done: add winner summary endpoints.
12. Done: add host/player snapshot endpoints for reconnect and screen hydration.
13. Done: add event outbox storage for committed gameplay events.
14. Done: add SSE endpoint for ordered outbox events and snapshot refetching.
15. Done: add a 50-player load helper for realtime connections and gameplay bursts.
16. Done: add Entra-ready auth seams before production login is wired.
17. Done: add player reconnect/heartbeat state tracking with committed connection events for meaningful transitions.
18. Next: smoke-test the full local backend flow, then wire the frontend to these backend contracts.

## Frontend Milestone Order

1. Point the demo app at the Go backend through environment-based API configuration.
2. Replace mock host game creation with backend-backed game runs.
3. Replace mock lobby/player data with backend join/rejoin APIs.
4. Replace mock cards with backend-assigned persisted cards.
5. Wire host call controls to backend called-word APIs.
6. Wire player marks and Bingo claims to backend APIs.
7. Show claim result and top 3 winners from backend state.
8. Add SSE for live state updates with polling/reconnect fallback.
9. Keep visual polish focused on clarity, not a redesign.

## First API Endpoints After Scaffold

Current backend endpoints:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Process health. |
| `GET` | `/readyz` | Dependency readiness. |
| `GET` | `/api/v1/version` | Service version and environment. |
| `POST` | `/api/v1/games` | Create a production V1 game run. |
| `GET` | `/api/v1/games/{gameID}` | Fetch game run state. |
| `POST` | `/api/v1/games/{gameID}/allowed-players` | Add an allowed player. |
| `GET` | `/api/v1/games/{gameID}/allowed-players` | List allowed players. |
| `POST` | `/api/v1/games/{gameID}/players` | Join or rejoin a player. |
| `POST` | `/api/v1/games/{gameID}/players/{playerID}/card` | Assign or return a persisted card. |
| `GET` | `/api/v1/games/{gameID}/players/{playerID}/card` | Fetch assigned card. |
| `POST` | `/api/v1/games/{gameID}/start` | Start the game. |
| `POST` | `/api/v1/games/{gameID}/pause` | Pause the game. |
| `POST` | `/api/v1/games/{gameID}/resume` | Resume the game. |
| `POST` | `/api/v1/games/{gameID}/finish` | Finish the game. |
| `POST` | `/api/v1/games/{gameID}/cancel` | Cancel the game. |
| `POST` | `/api/v1/games/{gameID}/calls` | Record the next called word. |
| `GET` | `/api/v1/games/{gameID}/calls` | List called words. |
| `PATCH` | `/api/v1/games/{gameID}/players/{playerID}/card/cells/{cellID}` | Mark or unmark a player card cell. |
| `POST` | `/api/v1/games/{gameID}/claims` | Submit and validate a Bingo claim. |
| `GET` | `/api/v1/games/{gameID}/claims` | List host claim state. |
| `GET` | `/api/v1/games/{gameID}/summary` | Return winners and final state. |
| `GET` | `/api/v1/games/{gameID}/host-snapshot` | Return host hydration/reconnect state. |
| `GET` | `/api/v1/games/{gameID}/players/{playerID}/snapshot` | Return player hydration/reconnect state. |
| `POST` | `/api/v1/games/{gameID}/players/{playerID}/heartbeat` | Refresh player online state and `last_seen_at`. |
| `GET` | `/api/v1/games/{gameID}/events` | Stream committed game events with SSE. |

Use `net/http` until route complexity proves a small router is worth adding.

## Realtime Backbone Status

The realtime backbone is implemented in the Go backend as a Postgres-backed event outbox plus snapshot-first SSE delivery.

- `game_event_outbox` stores committed gameplay events with `id`, `game_run_id`, `type`, `entity_id`, JSONB `payload`, per-game `sequence`, and `created_at`.
- Mutations write outbox rows in the same database transactions as game creation, lifecycle transitions, player joins, card assignment, marks, calls, claim validation, winners, and third-winner game finish.
- `GET /api/v1/games/{gameID}/events` streams outbox rows in sequence order with Server-Sent Events, supports `Last-Event-ID` and `?lastEventId=`, sends heartbeat comments, and closes cleanly on request cancellation.
- Payloads stay small; host and player screens should refetch snapshots after important events instead of treating SSE as the source of truth.
- `cmd/realtime-loadtest` can simulate at least 50 player/SSE connections, word calls, mark bursts, claim submissions, and reconnect snapshot fetches against a running local API.

Redis and Gorilla/WebSocket remain deferred on purpose. The current backend should first prove the single-instance, Postgres-authoritative path under local 50-player testing. Redis, Service Bus, Azure SignalR, or WebSocket fanout can be added after a measured scaling need appears or when multi-instance Azure deployment requires cross-process event delivery.

## Auth And Reconnect Hardening Status

The backend is Entra-ready but not yet running real Microsoft login.

- `AUTH_MODE=dev` remains the default local/test mode and maps development headers into the same internal principal used by services.
- `AUTH_MODE=entra-ready` switches to the same handler-facing `auth.Authenticator` boundary with bearer-token parsing, a token verifier interface, claims-to-principal mapping, and role mapping.
- Entra placeholders exist for tenant ID, client ID/audience, issuer, and JWKS URL. The default verifier is intentionally unconfigured, so local runs do not require Azure credentials, live Microsoft network calls, or JWKS fetching.
- Future real Entra work should implement the verifier behind the existing interface, validate issuer/audience/tenant/JWKS, and preserve the current `auth.Principal` shape.

Player connection state is now persisted and visible:

- Join creates players as `online`; rejoin updates the existing row to `online`, refreshes `last_seen_at`, and writes a committed `player.reconnected` event.
- Player snapshot fetches and `POST /api/v1/games/{gameID}/players/{playerID}/heartbeat` refresh `last_seen_at` for the authorized player or for host/admin acting on that player.
- Host snapshots include each player's `connectionState` and `lastSeenAt`.
- Heartbeat/snapshot refreshes do not emit noisy outbox rows while a player is already online, but reconnect transitions from offline/disconnected emit `player.reconnected`.
- The current `/events` stream is game-level; it does not infer per-player disconnects from SSE close. A later frontend should call the heartbeat endpoint while the player screen is open. A worker/timer-based timeout model can later mark stale players offline/disconnected if product needs exact presence.

## Schema Phases

### Phase 1: Production V1 Game Engine Tables

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

## Deferred Behind Interfaces

These are not part of the immediate realtime game backbone, but they are production V1 tracks and should stay behind interfaces instead of being hard-coded into handlers:

- Microsoft Entra ID production authentication.
- Real Microsoft Entra login, JWKS fetching, and JWT validation.
- Microsoft Graph Outlook or Teams delivery.
- Azure OpenAI content generation.
- Azure Speech voice calling.
- Gift card or voucher fulfillment.
- Real voice cloning or employee voice profiles.
- Azure deployment and managed identities.
- Durable Azure Service Bus workflows.
- Redis fanout.
- Gorilla/WebSocket realtime transport.
- Full admin console.
- Frontend wiring.

Use local development placeholders where needed. For example, development auth can identify the current user through headers or a local selector, generated word banks can be static, and winner notifications can be visible in the app instead of emailed.

## Immediate Next Backend Task

The next backend-only step is contract hardening after smoke verification: tighten authorization around remaining player-owned write paths such as card assignment, card fetch, mark/unmark, and claim submission, then add any missing API contract tests needed before frontend wiring. Keep Teams/Graph/email, AI caller, Azure Speech, rewards, Azure deployment, Redis, Gorilla/WebSocket, and real Entra/JWKS validation deferred until the local game loop is wired and smoke-tested end to end.
