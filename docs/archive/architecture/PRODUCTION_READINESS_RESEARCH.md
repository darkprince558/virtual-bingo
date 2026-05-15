# Virtual Bingo Production Readiness Research

Last updated: 2026-05-12

## Current State

The app is a polished frontend prototype, not a production system yet.

Evidence from the current codebase:

- Routes exist for landing, lobby, player game, host dashboard, and summary in the demo app: `apps/frontend-demo/app`.
- All gameplay state is mock/static: `apps/frontend-demo/lib/mockGameData.ts`.
- Host controls update only local UI state or `console.log`: `apps/frontend-demo/app/host/page.tsx`.
- Player claim opens the winner modal immediately instead of submitting a claim for validation: `apps/frontend-demo/app/play/page.tsx`.
- No API routes, database layer, authentication middleware, realtime server, audit export service, CI config, or tests exist yet.
- `npm run lint` passes.
- `npm run build` passes from `apps/frontend-demo`, with linting enabled during build after the workspace cleanup.
- `npm audit --json` reports 2 moderate vulnerabilities through `next -> postcss`.
- The earlier duplicate root lockfile was archived under `archive/generated/root-package-lock.json`; the demo app's active lockfile is `apps/frontend-demo/package-lock.json`.

## Production Goal

For a final full-scale production-ready site, Virtual Bingo should become a secure, authenticated, realtime event platform for internal workplace games:

- Hosts create and manage sessions.
- Players join assigned sessions with a game code.
- Cards are uniquely generated, persisted, and auditable.
- Called words are synchronized to all players in realtime.
- Bingo claims are server-validated before winners are confirmed.
- Results, claim history, and audit logs are exportable.
- The app can survive reconnects, host refreshes, concurrent players, and deployment scaling.

## P0: Make It A Real App

These are the non-negotiables before any serious production demo.

1. Replace mock state with a backend domain model.
   - Tables/entities: users, organizations/tenants, game sessions, players, cards, card cells, word banks, called words, claims, winners, audit events, prize notifications.
   - Preserve immutable event history for word calls, claims, approvals, rejections, and game-end summary.
   - Keep card assignment deterministic/auditable once a player joins.

2. Add authentication and authorization.
   - The BRD/design notes mention Microsoft Entra ID, so production should use Entra-backed sign-in.
   - Roles needed: host/admin, player, viewer/auditor.
   - Hosts should only manage sessions they own or are assigned to.
   - Players should only see their own card plus shared session state.

3. Build session APIs.
   - Create session.
   - Join session by code.
   - Start/pause/resume/end session.
   - Call next word.
   - Submit claim.
   - Approve/reject claim.
   - Fetch live session snapshot.
   - Fetch final summary/audit export.

4. Add realtime state sync.
   - Player events: session status, current word, called-word history, player count, claim state, leaderboard, winner confirmation.
   - Host events: joins/leaves, claims, connection state, audit events.
   - At scale, Socket.IO requires either sticky sessions when long-polling is enabled or WebSocket-only transport, and multi-node broadcasts need a compatible adapter such as Redis.

5. Implement server-side winner validation.
   - Client-side marks are not proof of bingo.
   - Validation must check the assigned card, marked cells, called words, and allowed winning patterns.
   - Claims should have clear states: pending, valid, invalid, confirmed, rejected.
   - Store validation evidence for audit/debugging.

## P1: Make It Safe, Operable, And Trustworthy

1. Security baseline.
   - Add security headers in `next.config.ts`: HSTS, frame protection/CSP `frame-ancestors`, content type nosniff, referrer policy, permissions policy.
   - Add CSRF protection or same-site session strategy for mutating endpoints.
   - Validate all server inputs with schemas.
   - Rate-limit join, claim, and host mutation endpoints.
- Keep unused env/dependency surface out of the demo and future production app. The previous Gemini/AI Studio scaffold was removed during the workspace cleanup.

2. Data correctness.
   - Use transactions for host actions that mutate game state.
   - Prevent duplicate word calls.
   - Prevent multiple active sessions with the same public code.
   - Prevent duplicate player joins unless reconnecting the same identity.
   - Ensure claim approval cannot race with game end.

3. Reconnect and failure handling.
   - Rehydrate player card and marks after refresh.
   - Restore host dashboard state after refresh.
   - Show degraded/reconnecting state in the UI.
   - Persist audit events even if realtime delivery fails.

4. Admin/audit/export.
   - Export final results as CSV and/or PDF.
   - Export full audit log with timestamps, actor, action, and relevant entity IDs.
   - Add a session summary page backed by persisted data.
   - Add prize notification status as a real workflow or explicitly keep it manual.

5. Observability.
   - Structured logs for session lifecycle and claims.
   - Error reporting.
   - Basic metrics: active sessions, connected clients, event latency, failed claims, reconnects.
   - Health endpoint for platform checks.

## P2: Make It Full Scale

1. Deployment architecture.
   - Next.js can deploy as Node.js server, Docker container, static export, or platform adapter; this app will need a Node/server deployment because it needs authenticated APIs and realtime/session behavior.
   - If using WebSockets, verify the hosting platform supports persistent connections and the chosen scaling model.
   - Containerize or platform-deploy with explicit environment variables, health checks, and release rollback.

2. Realtime scaling plan.
   - Single instance is fine for early internal events.
   - Multi-instance needs Redis/pub-sub or another adapter so events reach all connected clients.
   - Sticky session behavior matters if Socket.IO long-polling fallback is enabled.
   - Consider managed realtime alternatives if the team wants less infrastructure ownership.

3. Test strategy.
   - Unit tests for card generation, word calling, pattern validation, claim validation, scoring, and leaderboard.
   - API/integration tests for session lifecycle.
   - Playwright E2E for host creates game -> player joins -> host calls words -> player claims -> host approves -> summary exports.
   - Load test a realistic event: 50, 100, 250, 500 concurrent players depending on target.

4. CI/CD.
   - Required checks: install, lint, typecheck, build, unit tests, E2E smoke.
   - Dependency audit gate with an allowlist policy.
   - Preview deployments per PR.
   - Production deploy gate/manual approval for internal corporate environments.

5. Product finishing.
   - Host setup flow: word bank selection, pattern rules, prize settings, session name/date.
   - Player identity flow: authenticated join or guest policy.
   - Accessibility pass: keyboard operation, focus states, screen-reader labels, reduced motion.
   - Mobile/tablet pass for players.
   - Empty/loading/error states for every screen.

## Suggested Architecture

Recommended first production architecture:

- Frontend: current Next.js app.
- Backend: Next.js route handlers/server actions for normal CRUD plus a dedicated realtime server if the platform does not support WebSockets cleanly inside the Next process.
- Database: Postgres.
- Realtime fanout: Socket.IO or managed realtime service.
- Pub/sub/cache: Redis if using multi-node Socket.IO or rate limiting.
- Auth: Microsoft Entra ID / MSAL or a Next-compatible auth layer configured for Entra.
- Storage/export: generated CSV/PDF from persisted session data.
- Deployment: Node.js or Docker deployment, not static export.

## Immediate Implementation Order

1. Clean project foundations.
   - Remove or resolve duplicate root/package lockfile ambiguity.
   - Rename package from `ai-studio-applet` to the real app name.
   - Remove unused Gemini dependency/env docs unless AI caller generation is actually being built now.
   - Turn build linting back on.
   - Fix the moderate dependency audit issue by updating Next/PostCSS once verified.

2. Add backend data model and migrations.
   - Start with sessions, players, cards, called words, claims, audit events.
   - Use a seed script to reproduce the current mock demo data.

3. Implement backend-backed host/session flow.
   - Create/join session APIs.
   - Host call-next-word API.
   - Persist audit events.
   - Update existing UI to read from backend snapshots.

4. Implement claim validation.
   - Move bingo pattern detection to shared/server-tested domain logic.
   - Player submits claim.
   - Host reviews validation result.
   - Winners persist into summary.

5. Add realtime.
   - Broadcast state changes.
   - Reconnect and rehydrate.
   - Add connection status that reflects real socket state.

6. Add production hardening.
   - Auth/roles.
   - Rate limits.
   - Security headers.
   - Observability.
   - Export.
   - CI and E2E tests.

## Open Product Decisions

- Is this strictly internal to one company tenant, or should it support many organizations?
- Are players required to authenticate with Microsoft Entra, or can game-code guests join?
- Should hosts be able to customize word banks, winning patterns, and prize text?
- What is the target event size: 25, 100, 500, or more concurrent players?
- Should AI caller messages be generated live, pre-generated, or removed for v1?
- What export format is required by managers: CSV, PDF, email summary, or all three?
- Production direction is Azure, with exact service split covered in `AUTONOMOUS_BACKEND_ARCHITECTURE.md`.

## External References Checked

- Next.js deployment docs: https://nextjs.org/docs/app/getting-started/deploying
- Next.js security/custom headers docs: https://nextjs.org/docs/app/api-reference/config/next-config-js/headers
- Next.js Playwright testing docs: https://nextjs.org/docs/app/guides/testing/playwright
- Microsoft MSAL / Entra authentication docs: https://learn.microsoft.com/en-us/entra/msal/
- Microsoft identity platform auth flows: https://learn.microsoft.com/en-us/entra/identity-platform/authentication-flows-app-scenarios
- Socket.IO multi-node deployment docs: https://socket.io/docs/v4/using-multiple-nodes/
- Socket.IO Redis adapter docs: https://socket.io/docs/v4/redis-adapter/
