# Virtual Bingo Full-Scale Deployment Roadmap

Last updated: 2026-05-15

This document captures the larger product direction beyond the current Production V1 gameplay foundation. The goal is not only a locally playable bingo app. The goal is a full CGI-internal Microsoft-integrated cloud platform with real auth, scalable realtime play, autonomous weekly game operations, Microsoft 365 delivery, AI content, AI voice calling, operational auditability, and Azure deployment.

## Guiding Principle

Postgres remains the authoritative source of truth. Every other system is a delivery, automation, cache, fanout, or integration layer.

- Postgres owns durable game state, player state, cards, calls, claims, winners, event outbox, schedules, audit history, AI review state, and reward state.
- Redis/WebSockets/SSE deliver realtime updates but do not decide game truth.
- Microsoft Graph sends messages but does not own game eligibility or results.
- AI generates content/caller output but does not bypass host/admin controls or deterministic game validation.
- Azure Speech renders caller audio but does not own caller sequence.
- Service Bus/workers run automation but commit important state changes back to Postgres.

## Target End State

The full platform should support:

- CGI-only Microsoft login through a single trusted Microsoft Entra tenant.
- Roles for `admin`, `host`, and `player`.
- Admin-managed host access.
- Host-created game templates and manual/adaptive game runs.
- Allowlisted players per game.
- Realtime game play for at least 50 players first, then a path to 100+.
- Redis-backed realtime fanout for multi-instance deployment.
- SSE support and an intentional WebSocket transport when richer bidirectional realtime is needed.
- Weekly scheduled games.
- Microsoft Graph email invites, reminders, winner emails, and host summaries.
- Teams notification support after email delivery is stable.
- AI-generated bingo content with host review and auto-approval fallback rules.
- AI caller orchestration driven by committed game events.
- Azure Speech text-to-speech for caller audio.
- Rewards provider adapter and winner fulfillment state.
- Admin-visible game history and audit trails.
- Azure deployment with managed secrets, logs, metrics, and migration runbooks.

## Build Order

### 1. Cloud Foundation

Set up the Azure shape everything else plugs into.

- Azure hosting target, likely Azure Container Apps for the Go API and workers.
- Azure Database for PostgreSQL.
- Azure Cache for Redis.
- Azure Service Bus.
- Azure Key Vault.
- Application Insights or equivalent logs/metrics/traces.
- Container build and deployment workflow.
- Runtime environment separation: local, staging, production.
- Migration and rollback runbooks.

Why first: auth, jobs, Redis fanout, Graph delivery, Speech, and AI all need reliable cloud config and secret handling.

### 2. CGI Microsoft Auth And Roles

Move from dev auth / Entra-ready seams to real CGI-only Microsoft Entra authentication.

- Trust one CGI Microsoft Entra tenant for V1.
- Validate Microsoft JWTs with issuer, audience/client ID, tenant ID, and JWKS.
- Preserve the internal principal shape used by handlers/services.
- Enforce `admin`, `host`, and `player` behavior.
- Admins can grant/revoke host access.
- Hosts can create and manage only their own games.
- Players can access only games where they are allowlisted.

Why second: all production features depend on knowing who the user is and what they can do.

### 3. Authorization Hardening And Presence

Tighten ownership and player-specific access before wiring every feature.

- Harden player-owned paths: card assignment/fetch, marks, claims, heartbeat, snapshots.
- Keep heartbeat-based online/offline tracking.
- Add timeout-based offline detection through a worker or periodic cleanup.
- Show connection state in host snapshots.
- Avoid relying on raw SSE disconnect as the only presence signal.

Why third: production frontend wiring and live play need reliable reconnect behavior and clear access boundaries.

### 4. Frontend Wiring To The Go Backend

Turn the existing demo into a real backend-backed product flow.

- Replace mock host game creation with API calls.
- Replace mock player join/rejoin with allowlist-backed backend joins.
- Use persisted backend cards.
- Wire marks, calls, claims, winners, and summaries.
- Subscribe to SSE first and refetch snapshots after important events.
- Add heartbeat calls from player screens.
- Keep visual design workplace-appropriate and close to the current demo unless there is a specific product reason to redesign.

Why fourth: this is where the platform becomes usable end to end.

### 5. Redis Realtime Fanout And WebSocket Transport

Scale realtime delivery after the Postgres/SSE path is proven.

- Keep `game_event_outbox` as durable source.
- Add Redis pub/sub or streams for low-latency fanout across API instances.
- Keep SSE as the simple reliable client transport.
- Add WebSocket only when there is a clear bidirectional realtime need, such as host/caller control channels, richer presence, or lower-latency game interactions.
- Make reconnect behavior use snapshots plus event sequence resume.

Why fifth: Redis and WebSockets should strengthen a working event pipeline, not replace the source of truth.

### 6. Weekly Scheduling And Automation Workers

Introduce recurring operations.

- Game templates with recurrence, timezone, audience source, word set settings, winning pattern rules, and automation settings.
- Worker creates upcoming game runs.
- Worker snapshots audience/allowlist.
- Auto-start rules, including minimum player threshold.
- Job status and retry tracking.
- Audit trail for automation actions.

Why sixth: weekly automation needs auth, durable game state, and delivery channels to be useful.

### 7. Microsoft Graph Email And Teams Delivery

Automate player and host communication.

- Graph email invites.
- Reminder emails.
- Winner emails.
- Host summary emails.
- Delivery logs and retry statuses.
- Teams notification support after email works.

Why seventh: messages should be driven by scheduled game state and committed events, not ad hoc UI actions.

### 8. AI Content Generation

Add AI-generated game content with review controls.

- Prompt library.
- Generated word sets.
- Host review by default.
- Auto-approval fallback when configured.
- Content audit history.
- Reuse approved word sets.

Why eighth: AI content is valuable only when hosts can trust, review, and reuse it.

### 9. AI Caller And Azure Speech

Build the caller on top of the realtime/event pipeline.

- Caller state machine driven by committed game events.
- Call cadence and pause/resume behavior.
- Caller script generation.
- Azure Speech text-to-speech generation.
- Audio asset caching or streaming strategy.
- Voice selection.
- Admin-managed consent flow before any employee voice-profile work.

Why ninth: AI voice depends on reliable calls, pause/resume, event delivery, and host controls.

### 10. Rewards And Fulfillment

Add prizes after game results and audit trails are stable.

- Reward provider adapter.
- Winner fulfillment status.
- Admin-visible fulfillment history.
- Retry/manual resolution paths.
- Avoid storing raw gift card secrets unless the approved provider requires it.

Why tenth: rewards touch money and need careful auditability.

## What Not To Do

- Do not let Redis, WebSockets, AI, Speech, or Graph become the source of truth.
- Do not implement multi-company tenant support for V1; this is CGI-only.
- Do not bypass deterministic server-side bingo validation with AI.
- Do not hard-code one reward provider until the provider is confirmed.
- Do not build Teams/Graph/AI features before auth and durable game state can enforce access.
- Do not remove SSE just because WebSockets are added; SSE remains useful for simple ordered event delivery.

## Near-Term Priority

The next practical sequence is:

1. Authorization hardening for all remaining player-owned backend paths.
2. Frontend wiring to the Go backend.
3. Azure cloud foundation and real CGI Entra authentication.
4. Redis realtime fanout.
5. Weekly scheduler and Graph email automation.
6. AI content, AI caller, Azure Speech, Teams, and rewards.

This keeps the work aggressive toward the full platform while still building on stable dependencies.
