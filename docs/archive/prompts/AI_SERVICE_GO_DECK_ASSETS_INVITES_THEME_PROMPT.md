# AI Service Go Deck, Assets, Invites, Lobby, And Theme Prompt

Status: ready to use after `AI_SERVICE_GO_FOUNDATION_PROMPT.md` is completed.

```text
You are working in:

/Users/anish/Downloads/Work/BingoGame

Current branch:
codex/project-foundation-docs

Project direction:
Virtual Bingo is becoming a full-scale internal workplace bingo platform. The Go backend is the source of truth for scheduled game operations, access, game state, cards, called words, claims, winners, audit history, and realtime frontend events. The Python service owns AI work: topic generation, word list generation, caller lines, Azure Speech audio, and generated theme drafts.

This prompt is batch 2 of the Go AI/autonomous-game work. It assumes batch 1 already added:
- scheduler/automation foundation
- Python AI service client boundary
- generated content storage
- optional host review APIs
- content lock system

Goal:
Finish the rest of the scheduled AI-backed game-prep pipeline except lobby chat:
1. locked randomized call deck
2. caller asset/audio storage and Python bulk generation
3. invite/delivery job foundation with game-code links
4. lobby timing/state automation and icon/avatar selection storage
5. theme generation persistence and approval

Do not implement:
- frontend wiring
- real Microsoft Graph/Teams delivery
- real Azure deployment
- lobby chat
- reward fulfillment
- real voice-profile consent flows
- microphone claim recordings unless already present and tiny to connect

Read these first:
- README.md
- docs/README.md
- docs/architecture/AI_SERVICE_INTEGRATION_PLAN.md
- docs/architecture/V1_IMPLEMENTATION_PLAN.md
- docs/architecture/FULL_SCALE_DEPLOYMENT_ROADMAP.md
- docs/architecture/FULL_SCALE_FEATURE_API_BACKLOG.md
- docs/architecture/AUTONOMOUS_BACKEND_ARCHITECTURE.md
- backend-go/README.md
- backend-go/internal/config/config.go
- backend-go/internal/domain/types.go
- backend-go/internal/db/store.go
- backend-go/internal/game/service.go
- backend-go/internal/app/routes.go
- backend-go/internal/app/game_handlers.go
- backend-go/internal/app/management_handlers.go
- backend-go/internal/app/api.go
- backend-go/internal/app/api_test.go
- backend-go/internal/db/migrations/
- docs/archive/prompts/AI_SERVICE_GO_FOUNDATION_PROMPT.md

Hard constraints:
- Backend Go work only.
- Keep standard library net/http.
- Keep handlers thin.
- Keep AI/Python calls behind interfaces.
- Do not make Python or AI the source of truth.
- Go must own the final randomized call deck and live called-word state.
- Do not make the live game wait on Azure Speech under normal conditions.
- Do not import Azure SDKs into Go for this pass.
- Use local/mock delivery for emails/Teams; do not implement real Graph.
- Plain go test ./... must pass without TEST_DATABASE_URL.
- DB integration tests should skip when TEST_DATABASE_URL is not set.
- Preserve response envelope:
  - success: { "data": ... }
  - error: { "error": { "code": "...", "message": "..." } }

Implementation tasks:

1. Locked randomized call deck
- Add schema/store/domain/service support for a locked call deck.
- Suggested table: `game_call_deck`.
- Store:
  - id
  - game_run_id
  - word_set_word_id or word text
  - sequence
  - shuffle_seed
  - shuffle_version
  - locked_at
  - called_word_id nullable
  - created_at
- At content lock or a new explicit post-lock step, create the randomized deck from the locked word list.
- Randomization must be deterministic from the stored seed/version so it is auditable/reproducible.
- Host should review only the word list/topic/settings, not the final call order.
- Change live word calling so `POST /api/v1/games/{gameID}/calls` calls the next uncalled deck item when a locked deck exists.
- Preserve existing fallback behavior for games without a locked deck if needed for local demo compatibility.
- When a word is called, mark/link the deck item to the created called word.
- Add clear conflict errors when the deck is exhausted or missing for locked AI-backed games.

2. Caller asset/audio storage and Python bulk generation
- Extend the Python AI client boundary with:
  - GenerateCallerAssetsBulk(ctx, input) (output, error)
  - GenerateCallerAsset(ctx, input) (output, error) for fallback/regeneration
- Use service-to-service endpoints from docs/architecture/AI_SERVICE_INTEGRATION_PLAN.md:
  - POST /ai/v1/caller-assets/bulk
  - POST /ai/v1/caller-assets
- Add schema/store/domain/service support for caller assets.
- Suggested table: `caller_assets`.
- Store:
  - id
  - game_run_id
  - call_deck_item_id
  - word
  - sequence
  - line
  - audio_url or storage_key
  - voice_name
  - provider
  - status: pending, ready, failed, fallback
  - error_reason
  - created_at/updated_at
- Add a post-lock service method and optional manual endpoint:
  - POST /api/v1/games/{gameID}/caller-assets/generate
- It should send the full locked deck to Python, persist per-item results, and mark failures without blocking gameplay.
- Add fallback text for any failed/missing caller asset, for example `Next word is {word}.`
- Expose caller asset info where the future frontend can get it:
  - called-word response can include caller asset if ready
  - host/player snapshots can include current word caller line/audio metadata
  - SSE `word.called` payload can include compact caller metadata or clients can refetch snapshots
- Emit outbox/activity events:
  - `caller.assets_generation_started`
  - `caller.audio_ready`
  - `caller.failed`

3. Invite/delivery job foundation
- Add delivery data model without real Graph/Teams yet.
- Suggested tables:
  - delivery_batches
  - delivery_attempts
- Store:
  - game_run_id
  - channel: email, teams
  - purpose: host_review, player_invite, reminder, summary
  - recipient email/user id
  - subject/template key/body preview
  - link/code payload
  - status: pending, sent, failed, skipped
  - error_reason
  - created_at/updated_at/sent_at
- Add a local/mock delivery service interface so future Graph can plug in.
- Add manual/local endpoint if useful:
  - POST /api/v1/games/{gameID}/deliveries/player-invites
  - GET /api/v1/games/{gameID}/deliveries
  - POST /api/v1/deliveries/{deliveryID}/retry
- Player invite links can include the public code:
  - `/join?code=ABCD12`
- Login and allowlist checks still control access. The code only fills in lookup/join context.
- Delivery generation should use the existing allowed players list.
- Do not send real email yet; in local mode record mock sent attempts and body/link payload.

4. Lobby timing/state automation and icon/avatar selection
- Add explicit lobby-opening support for T minus 10.
- Reuse existing `lobby_open` status if present; add service method:
  - OpenLobby(ctx, gameRunID)
- Add scheduled/due-job path or manual endpoint:
  - POST /api/v1/games/{gameID}/lobby/open
- Enforce sensible lifecycle transitions:
  - draft/scheduled/invites_sent -> lobby_open
  - lobby_open -> live through existing start flow
  - avoid reopening finished/cancelled games
- Add icon/avatar selection storage for players.
- Suggested player fields or table:
  - player_icon
  - player_avatar_color
  - player_avatar_label
  - updated_at
- Add current-player API:
  - PATCH /api/v1/games/{gameID}/players/me/profile
  - maybe GET is included in existing snapshot/player response
- Validate avatars against a safe fixed set or simple constraints. Do not accept arbitrary image URLs.
- Add outbox event:
  - `player.profile_updated`
- Include lobby state and player icon/avatar in host/player snapshots.

5. Theme generation persistence and approval
- Extend Python AI client boundary with:
  - GenerateTheme(ctx, input) (output, error)
- Add schema/store/domain/service support.
- Suggested tables:
  - theme_generation_jobs
  - themes
  - theme_approvals
- AI-generated themes must be structured tokens only:
  - name
  - summary
  - palette
  - icons
  - decorations
  - motion
  - callerTone
  - accessibility metadata
- Reject arbitrary CSS, JavaScript, unapproved image URLs, or unsafe asset references.
- Add host/admin APIs:
  - POST /api/v1/themes/generate
  - GET /api/v1/themes/{themeID}
  - PATCH /api/v1/themes/{themeID}
  - POST /api/v1/themes/{themeID}/approve
  - POST /api/v1/themes/{themeID}/reject
  - POST /api/v1/games/{gameID}/theme
  - GET /api/v1/theme-assets
- Applying a theme should update existing game settings/theme fields if those exist.
- Add outbox events:
  - `theme.generated`
  - `theme.approved`
  - `theme.applied`
- Keep theme frontend rendering out of scope.

Tests required:
- Unit tests for deterministic deck randomization from seed/version.
- Unit/service tests for deck-based call order.
- AI client tests for bulk caller assets and theme generation with mock/noop behavior.
- DB/API integration tests with TEST_DATABASE_URL for:
  - locked deck creation
  - calls follow deck sequence
  - caller asset bulk generation persists ready/failed rows
  - snapshots/call responses expose caller metadata
  - mock player-invite delivery creates attempts from allowlist
  - lobby open lifecycle transition
  - player profile/avatar update and snapshot visibility
  - theme generate/approve/apply flow
- Plain go test ./... must pass without TEST_DATABASE_URL.

Docs required:
- Update backend-go/README.md with:
  - new env vars
  - deck/caller asset endpoints
  - mock delivery behavior
  - lobby/profile endpoints
  - theme endpoints
- Update docs/architecture/V1_IMPLEMENTATION_PLAN.md with implemented milestone status.
- Update docs/architecture/AI_SERVICE_INTEGRATION_PLAN.md if implementation route names/schema differ.
- Update docs/architecture/FULL_SCALE_FEATURE_API_BACKLOG.md if theme/caller/delivery API status changes.
- Keep docs honest about what remains deferred:
  - frontend wiring
  - real Graph/Teams delivery
  - Azure deployment
  - rewards
  - voice claim recordings
  - lobby chat

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

PORT=18081 APP_ENV=test DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable AI_SERVICE_ENABLED=false CORS_ALLOWED_ORIGINS=http://localhost:3000 go run ./cmd/api

Smoke with curl:
- create game
- add allowed players
- prepare/patch/lock content if batch 1 endpoints exist
- create locked randomized deck
- generate mock caller assets
- send mock player invites
- open lobby
- player joins and updates icon/avatar
- start game
- call words and verify deck sequence
- fetch host/player snapshots with caller/theme/profile data
- generate/approve/apply theme

Commit discipline:
Make small commits if possible:
1. call-deck schema/store/service
2. caller-assets client/schema/service
3. delivery batch/attempt foundation
4. lobby/profile state
5. theme generation/persistence APIs
6. tests and docs

End with:
- files changed
- commands run
- tests passed
- API smoke summary
- remaining backend gaps only
```
