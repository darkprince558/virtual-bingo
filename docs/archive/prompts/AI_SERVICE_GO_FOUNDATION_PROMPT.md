# AI Service Go Foundation Prompt

Status: ready to use for the next backend implementation chat.

```text
You are working in:

/Users/anish/Downloads/Work/BingoGame

Current branch:
codex/project-foundation-docs

Project direction:
Virtual Bingo is becoming a full-scale internal workplace bingo platform. The Go backend is the source of truth for schedules, players, access, cards, calls, claims, winners, audit history, and realtime frontend events. The Python service owns AI tasks: weekly topic generation, word list generation, caller line generation, Azure Speech audio, future theme drafts, and future recap copy.

Important product flow:
- T minus 60 minutes: Go starts game prep. Python generates topic + word list + summary. Go emails/records a host review package.
- T minus 60 to T minus 30: host review is optional. If the host edits content/settings, those edits apply.
- T minus 30: Go locks the final content and prevents edits that would invalidate later generation.
- After lock: later work will randomize the deck and ask Python to bulk-generate caller audio.
- T minus 10: later work opens lobby.
- Live game: Go calls words and validates claims. Python is not in the live critical path except fallback/regeneration.

This prompt is only batch 1 of the Go work. Do the first 5 missing Go pieces only:
1. scheduler/automation foundation for the T-60 and T-30 jobs
2. Python AI service client boundary
3. generated content storage
4. optional host review APIs
5. content lock system

Do not implement the next batch yet:
- locked randomized call deck
- caller asset/audio storage
- invite/email delivery integration
- lobby timing/state automation
- theme generation persistence
- lobby chat
- frontend wiring
- real Microsoft Graph/Teams delivery
- real Azure deployment

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

Current backend capabilities:
- Go API with net/http.
- Dev auth and Entra-ready auth seams.
- Postgres migrations and store layer.
- Game create/list/read/update.
- Allowlist management.
- Player join/rejoin, snapshots, heartbeat.
- Card assignment and marking.
- Game settings and player preferences.
- Auto-mark and claim-readiness.
- Start/pause/resume/finish/cancel.
- Called words and called-word history.
- Claim validation, top-3 winners, summary.
- Postgres event outbox and SSE.
- Response envelope contract:
  - success: { "data": ... }
  - error: { "error": { "code": "...", "message": "..." } }

Hard constraints:
- Backend Go work only.
- Keep standard library net/http.
- Keep handlers thin.
- Keep AI service integration behind an interface/client boundary.
- Do not import Azure SDKs into Go for this pass.
- Do not make Python or AI the source of game truth.
- Do not change deterministic bingo validation.
- Do not wire the frontend.
- Do not implement real email/Graph/Teams delivery yet.
- Do not implement caller audio/deck generation yet; only prepare the lock foundation.
- Plain go test ./... must pass without TEST_DATABASE_URL.
- DB integration tests should skip when TEST_DATABASE_URL is not set.
- Preserve existing response envelope, dev auth headers, request ID behavior, and CORS behavior.

Implementation tasks:

1. Scheduler / automation foundation
- Add a small internal automation/scheduler package or service boundary for due jobs.
- It should be local-process friendly and testable without sleeping in tests.
- Add service methods that can be called by a future worker/cron:
  - PrepareGameContent(ctx, gameRunID)
  - LockGameContent(ctx, gameRunID)
- Do not build a full distributed scheduler yet.
- It is acceptable to expose manual internal/admin HTTP endpoints for local testing, for example:
  - POST /api/v1/games/{gameID}/content/prepare
  - POST /api/v1/games/{gameID}/content/lock
- Require host/admin auth for manual endpoints.
- Document that real scheduled execution will later move to Azure Container Apps Jobs, Functions, or Service Bus workers.

2. Python AI service client boundary
- Add typed config for the Python AI service, for example:
  - AI_SERVICE_BASE_URL
  - AI_SERVICE_TIMEOUT_SECONDS
  - AI_SERVICE_ENABLED
- Add an internal client interface, for example:
  - GenerateGamePrep(ctx, input) (output, error)
- Add an HTTP implementation that calls POST /ai/v1/game-prep using the shape in docs/architecture/AI_SERVICE_INTEGRATION_PLAN.md.
- Add a mock/noop implementation for tests and local disabled mode.
- Validate/normalize returned words:
  - trim whitespace
  - reject blank words
  - dedupe case-insensitively
  - enforce a minimum useful word count, ideally at least 24 active words for card generation
- Do not call caller-assets endpoints in this batch.

3. Generated content storage
- Add migrations and store methods for generated game prep content.
- Keep the schema simple but future-proof.
- Suggested tables:
  - content_generation_jobs
  - generated_game_content
  - game_run_content_reviews
- Track:
  - game_run_id
  - job type/status/provider
  - topic
  - summary
  - generated words as JSONB or normalized rows
  - current editable/final words
  - host edits
  - review window opens/closes
  - locked_at
  - errors/retry count
  - created_at/updated_at
- Use Postgres as the source of truth.
- Add domain types for generated content and job status.

4. Optional host review APIs
- Add host/admin APIs to fetch and edit the generated review package before lock.
- Suggested routes:
  - GET /api/v1/games/{gameID}/content
  - PATCH /api/v1/games/{gameID}/content
- GET should return:
  - gameRunId
  - status: draft/generated/edited/locked/failed or similar
  - topic
  - summary
  - words
  - reviewWindowClosesAt
  - lockedAt
  - generation metadata/error if useful
- PATCH should allow pre-lock edits to:
  - topic
  - summary
  - words
  - callerStyle/tone placeholders if already useful
- PATCH must reject edits after lock.
- PATCH must validate the final word list and return clear validation errors.
- Write audit/outbox activity events for generation complete, host edits, and lock where appropriate.

5. Content lock system
- Add service/store logic to lock generated content for a game.
- Lock should:
  - freeze final topic/summary/words
  - set locked_at
  - block later edits
  - ensure the word list is valid for card generation
  - optionally create or update a manual word set from the locked words if that fits existing game/card flow
  - associate the game run with the locked word set if needed
  - record audit/outbox events
- Do not randomize/store the ordered call deck in this batch unless it falls out naturally and stays tiny. Prefer leaving deck randomization to the next batch.
- Do not generate caller audio in this batch.

Tests required:
- Unit tests for AI client disabled/mock behavior and response validation.
- Unit tests or service tests for pre-lock edit validation and post-lock edit rejection.
- DB/API integration tests with TEST_DATABASE_URL for:
  - prepare content creates a generation job/content row
  - GET content returns the generated package
  - PATCH content edits words before lock
  - lock freezes the content
  - PATCH after lock returns a clear error
  - locked words can be used by existing card assignment path, if you wire the locked content into word sets in this batch
- Plain go test ./... must still pass without TEST_DATABASE_URL.

Docs required:
- Update backend-go/README.md with:
  - new env vars
  - new content prep/review/lock endpoints
  - local disabled/mock AI behavior
- Update docs/architecture/V1_IMPLEMENTATION_PLAN.md with the new backend milestone status.
- Update docs/architecture/AI_SERVICE_INTEGRATION_PLAN.md if the implemented route names or schema differ from the plan.
- Keep docs honest about what is still deferred:
  - deck randomization storage
  - caller assets/audio
  - invites/email/Teams
  - lobby automation
  - theme generation
  - chat

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
- create game as host
- prepare content
- fetch content
- patch content words before lock
- lock content
- verify post-lock patch fails
- assign player/card using locked content if implemented

Commit discipline:
Make small commits if possible:
1. AI config/client boundary
2. content-generation migrations/store/domain
3. prepare/lock service methods
4. content HTTP handlers/routes
5. tests and docs

End with:
- files changed
- commands run
- tests passed
- API smoke summary
- remaining Go gaps for batch 2 only
```
