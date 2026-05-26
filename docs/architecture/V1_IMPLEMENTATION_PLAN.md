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
- Keeping the existing workplace styling useful without treating it as final production UI.

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
18. Done: add local management API contracts for current identity, config/capabilities, game lookup/list/update, allowlist management, word-set CRUD, player `/me` aliases, claim detail, and host activity reads.
19. Done: add backend feature discovery flags, durable game-run settings, player marking preferences, auto-mark mode, auto-mark backfill, and player claim-readiness.
20. Done: add the first Go automation/content foundation for T-60 prep and T-30 lock: Python AI client boundary, local disabled AI mode, generated content tables, host review APIs, lock APIs, audit/outbox events, and locked generated word sets for existing card assignment.
21. Done: add the second Go AI/autonomous-game batch: deterministic locked call deck storage, deck-based word calling, caller asset metadata storage with Python bulk/single client methods and local fallback lines, mock invite delivery attempts with game-code links, T-10 lobby open hook, player icon/avatar profile storage, and structured theme generation/approval/application.
22. Next: wire the frontend to the now-persisted backend flows, then add real Microsoft Graph/Teams/email delivery, Azure Speech storage/execution, and deployment workers in later batches.

## Frontend Milestone Order

1. Point the web app at the Go backend through environment-based API configuration.
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
| `GET` | `/api/v1/config` | Public local capability/config discovery for the frontend. |
| `GET` | `/api/v1/me` | Return/upsert the authenticated current user. |
| `POST` | `/api/v1/games` | Create a production V1 game run. |
| `GET` | `/api/v1/games` | List games by `host`, `player`, or admin-only `admin` scope with optional status filter. |
| `GET` | `/api/v1/games/{gameID}` | Fetch game run state. |
| `GET` | `/api/v1/games/code/{code}` | Fetch game run state by public code. |
| `PATCH` | `/api/v1/games/{gameID}` | Update editable pre-live game metadata. |
| `POST` | `/api/v1/games/{gameID}/allowed-players` | Add an allowed player. |
| `POST` | `/api/v1/games/{gameID}/allowed-players/bulk` | Bulk add allowlist rows all-or-nothing. |
| `GET` | `/api/v1/games/{gameID}/allowed-players` | List allowed players. |
| `PATCH` | `/api/v1/games/{gameID}/allowed-players/{allowedPlayerID}` | Update an allowlist row. |
| `DELETE` | `/api/v1/games/{gameID}/allowed-players/{allowedPlayerID}` | Delete an allowlist row inside the game. |
| `POST` | `/api/v1/games/{gameID}/players` | Join or rejoin a player. |
| `GET` | `/api/v1/games/{gameID}/players/me/snapshot` | Return current-player hydration/reconnect state by auth email. |
| `POST` | `/api/v1/games/{gameID}/players/me/card` | Assign or return the current player's persisted card. |
| `GET` | `/api/v1/games/{gameID}/players/me/card` | Fetch the current player's assigned card. |
| `PATCH` | `/api/v1/games/{gameID}/players/me/card/cells/{cellID}` | Mark or unmark a current-player card cell. |
| `POST` | `/api/v1/games/{gameID}/players/me/heartbeat` | Refresh current-player online state. |
| `POST` | `/api/v1/games/{gameID}/players/{playerID}/card` | Assign or return a persisted card. |
| `GET` | `/api/v1/games/{gameID}/players/{playerID}/card` | Fetch assigned card. |
| `POST` | `/api/v1/games/{gameID}/start` | Start the game. |
| `POST` | `/api/v1/games/{gameID}/pause` | Pause the game. |
| `POST` | `/api/v1/games/{gameID}/resume` | Resume the game. |
| `POST` | `/api/v1/games/{gameID}/finish` | Finish the game. |
| `POST` | `/api/v1/games/{gameID}/cancel` | Cancel the game. |
| `GET` | `/api/v1/games/{gameID}/settings` | Host/admin read of lazily created durable game settings. |
| `PATCH` | `/api/v1/games/{gameID}/settings` | Host/admin patch of marking, claim-readiness, voice placeholder, caller placeholder, and theme placeholder settings. |
| `POST` | `/api/v1/games/{gameID}/content/prepare` | Host/admin manual hook for the T-60 generated content prep job. |
| `GET` | `/api/v1/games/{gameID}/content` | Host/admin read of generated topic, summary, words, review window, lock state, and provider metadata. |
| `PATCH` | `/api/v1/games/{gameID}/content` | Host/admin pre-lock edits to generated topic, summary, words, and caller style placeholder. |
| `POST` | `/api/v1/games/{gameID}/content/lock` | Host/admin manual hook for the T-30 content lock job; creates/attaches an approved generated word set. |
| `GET` | `/api/v1/games/{gameID}/players/me/preferences` | Current-player read of optional marking preferences. |
| `PATCH` | `/api/v1/games/{gameID}/players/me/preferences` | Current-player patch of optional marking mode when host allows player choice. |
| `POST` | `/api/v1/games/{gameID}/auto-mark/run` | Host/admin idempotent backfill for already-called words and effective `auto_mark` players. |
| `GET` | `/api/v1/games/{gameID}/players/me/claim-readiness` | Current-player UX helper for supported claim readiness from persisted card/call state. |
| `POST` | `/api/v1/games/{gameID}/calls` | Record the next called word. |
| `GET` | `/api/v1/games/{gameID}/calls` | List called words. |
| `PATCH` | `/api/v1/games/{gameID}/players/{playerID}/card/cells/{cellID}` | Mark or unmark a player card cell. |
| `POST` | `/api/v1/games/{gameID}/claims` | Submit and validate a Bingo claim. |
| `GET` | `/api/v1/games/{gameID}/claims` | List host claim state. |
| `GET` | `/api/v1/games/{gameID}/claims/{claimID}` | Read claim detail when host/admin or owning player. |
| `GET` | `/api/v1/games/{gameID}/summary` | Return winners and final state. |
| `GET` | `/api/v1/games/{gameID}/host-snapshot` | Return host hydration/reconnect state. |
| `GET` | `/api/v1/games/{gameID}/players/{playerID}/snapshot` | Return player hydration/reconnect state. |
| `POST` | `/api/v1/games/{gameID}/players/{playerID}/heartbeat` | Refresh player online state and `last_seen_at`. |
| `GET` | `/api/v1/games/{gameID}/events` | Stream committed game events with SSE. |
| `GET` | `/api/v1/games/{gameID}/activity` | Return committed host/admin activity feed events. |
| `GET` | `/api/v1/word-sets` | List authenticated-readable word sets. |
| `POST` | `/api/v1/word-sets` | Create a manual/seed word set. |
| `GET` | `/api/v1/word-sets/{wordSetID}` | Read word set detail and words. |
| `PATCH` | `/api/v1/word-sets/{wordSetID}` | Update word set metadata/status. |
| `POST` | `/api/v1/word-sets/{wordSetID}/words` | Add a manual word. |
| `PATCH` | `/api/v1/word-sets/{wordSetID}/words/{wordID}` | Update a manual word. |
| `DELETE` | `/api/v1/word-sets/{wordSetID}/words/{wordID}` | Soft-deactivate a word. |

Use `net/http` until route complexity proves a small router is worth adding.

### Local Management API Status

This backend slice deliberately stays local and backend-only. It does not add Azure, Microsoft Graph, rewards, real automation jobs, voice-claim recordings, lobby chat, or frontend screens.

Deliberate choices:

- `GET /api/v1/config` is public because it only exposes local capability flags and reconnect timing. It now reports `gameSettings=true`, `autoMark=true`, `aiContent=true`, `aiCaller=true`, `themeGenerator=true`, and `automation=true`, while voice claims, Teams app, and rewards stay false.
- Bulk allowlist insert is all-or-nothing. Duplicate emails, missing fields, or database conflicts reject the whole request.
- Game code edits are allowed only before live gameplay states. The code is normalized to uppercase and still goes through the database uniqueness constraint.
- Word deletion is soft deletion through `isActive=false`; gameplay card assignment still uses only active words and still requires at least 24 active words.
- Activity feed reads from the committed game event outbox for now. Actor and entity-type data can be enriched later from audit rows if the host UI needs it.
- Game settings live in `game_run_settings` with defaults created lazily. Player marking preferences live in `player_preferences`; game settings win unless `allow_player_marking_mode_choice=true`.
- Auto-mark is real backend behavior now. `manual` preserves manual marking, `assist` does not mark automatically, and `auto_mark` marks matching unmarked cells transactionally with called-word commits or through the idempotent backfill endpoint.
- Voice claim settings are stored as explicit placeholders only. Generated content prep/review/lock, caller asset metadata, structured theme persistence, mock invite delivery, and manual automation hooks are implemented, but no microphone upload, Azure Speech execution, real Teams/email behavior, rewards, cloud deployment, or frontend wiring has been added.

### AI Content Prep/Review/Lock Status

The first Go-side content automation slice is implemented.

- `AI_SERVICE_ENABLED`, `AI_SERVICE_BASE_URL`, and `AI_SERVICE_TIMEOUT_SECONDS` configure the Python AI service boundary.
- The Go client calls Python `POST /ai/v1/game-prep` for draft topic, summary, words, caller style, and theme prompt placeholders.
- Local disabled mode returns deterministic placeholder content so tests and smoke flows do not need Python or Azure.
- `content_generation_jobs`, `generated_game_content`, and `game_run_content_reviews` persist generation status, provider metadata, generated words, current edited words, review window timestamps, lock timestamp, and host edits.
- Manual host/admin endpoints exist for local testing the future T-60 prep and T-30 lock jobs.
- Lock freezes the final content, blocks later edits, creates an approved `ai_generated` word set, and associates the game run with that word set so existing card assignment can continue through the deterministic word-set path.

The follow-on autonomous game-prep slice is now implemented too.

- `game_call_deck` stores the locked post-content randomization using a stored seed and shuffle version.
- Live `POST /api/v1/games/{gameID}/calls` follows the locked deck when present and preserves the older active-word fallback for local development games without a deck.
- `caller_assets` stores deck-linked caller lines, provider/status, optional audio URL/storage key, and fallback text so live gameplay does not wait on Azure Speech.
- `delivery_batches` and `delivery_attempts` store local/mock player-invite delivery attempts with `/join?code={CODE}` links; real Graph/Teams/email delivery remains behind the delivery boundary.
- `POST /api/v1/games/{gameID}/lobby/open` is the manual T-10 hook, and player avatar profile fields are persisted on `players`.
- `theme_generation_jobs`, `themes`, and `theme_approvals` store structured AI theme drafts, host edits, approval/rejection, and game settings application.

Still deferred from this track: frontend wiring, real Graph/Teams/email delivery, Azure Speech execution/storage integration, Azure deployment, rewards, voice claim recordings/voice consent flows, and lobby chat.

## Realtime Backbone Status

The realtime backbone is implemented in the Go backend as a Postgres-backed event outbox plus snapshot-first SSE delivery.

- `game_event_outbox` stores committed gameplay events with `id`, `game_run_id`, `type`, `entity_id`, JSONB `payload`, per-game `sequence`, and `created_at`.
- Mutations write outbox rows in the same database transactions as game creation, lifecycle transitions, player joins, card assignment, marks, calls, claim validation, winners, and third-winner game finish.
- Auto-mark writes `card.auto_marked` outbox rows with compact counts so clients can refetch snapshots instead of replaying per-cell deltas.
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
- The API runs a configurable stale-player sweeper using `PLAYER_CONNECTION_TIMEOUT_SECONDS`, `PLAYER_CONNECTION_SWEEP_INTERVAL_SECONDS`, and `PLAYER_CONNECTION_SWEEP_BATCH_SIZE`. Stale online players in active lobby/live/paused games move to `disconnected` and emit `player.disconnected`.
- When a disconnected player returns through rejoin, heartbeat, or player snapshot, responses can include `reconnectNotice.missedCalledWords`, which contains called words after the player's previous `lastSeenAt`. This gives the future frontend the exact data needed for a "while you were away" notification.
- The current `/events` stream is game-level; it does not infer per-player disconnects from SSE close. The frontend should keep calling heartbeat while the player screen is open and use `reconnectNotice` on return.

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
- `content_generation_jobs` - implemented for game prep jobs.
- `generated_game_content` - implemented for draft/current/locked content.
- `game_run_content_reviews` - implemented for host edit history.
- `generated_word_sets` - represented for this slice by approved `ai_generated` `word_sets` created at lock.
- `game_call_deck` - implemented for deterministic locked call order.
- `caller_assets` - implemented for generated/fallback caller line and audio metadata.
- `delivery_batches` - implemented for local/mock delivery grouping.
- `delivery_attempts` - implemented for local/mock invite attempts and retry records.
- `theme_generation_jobs` - implemented for AI theme draft jobs.
- `themes` - implemented for structured theme tokens and approval status.
- `theme_approvals` - implemented for host approval/rejection history.
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

The next backend-only batch should avoid redoing the autonomous game-prep foundation. Remaining backend gaps are real delivery providers, Azure Speech/audio object storage integration, production Entra/JWKS validation, deployment worker wiring for the manual automation hooks, and later rewards/voice consent flows. Lobby chat and frontend wiring remain separate batches.
