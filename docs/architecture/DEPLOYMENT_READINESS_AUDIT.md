# Virtual Bingo Deployment Readiness Audit

Date: 2026-05-26

This is the active deployment-start checklist for the cleaned repository. It intentionally focuses on what a team needs to deploy the real app, not the old presentation flow.

## Current State

The repo now has three active application surfaces:

- `apps/web` - Next.js web app.
- `backend-go` - Go API, migrations, game engine, SSE, and local development auth.
- `ai-service` - Python AI/content and caller-asset service.

The removed presentation/demo surfaces were:

- `/demo` route.
- `NEXT_PUBLIC_PRESENTATION_MOCK_MODE`.
- `presentationMockApi.ts` and `presentationMockMode.ts`.
- `scripts/demo-backend.sh`, `scripts/demo-frontend.sh`, and `scripts/demo-smoke.sh`.
- old archive/generated artifacts and obsolete prompt/demo docs.

## Do Not Ship Yet

Do not expose this to employees as production until these are done:

- Real Microsoft Entra JWT verification.
- Frontend auth/session wiring that sends bearer tokens instead of dev headers.
- Hosted Postgres with backups and migration runbook.
- Backend and AI service container images built in CI.
- Frontend deployment with `NEXT_PUBLIC_API_URL` pointing at the hosted Go API.
- CORS restricted to the hosted frontend origin.
- Basic rate limits on join, claim, content generation, caller generation, and delivery endpoints.
- Observability for API errors, DB readiness, and frontend runtime errors.

## P0 Deployment Work

1. Keep CI green and required.
   - `.github/workflows/ci.yml` now checks the Go backend, Postgres-backed Go tests, the web app, and the AI service.
   - Make this workflow a required branch protection check before multiple people start deployment work.

2. Package deployable services.
   - `backend-go/Dockerfile` builds the API and migration binary.
   - Keep `ai-service/Dockerfile`, but replace Flask's development server before pilot production.
   - Decide whether `apps/web` deploys as a hosted Next app or from its `output: 'standalone'` build.

3. Provision staging.
   - Managed Postgres.
   - API host.
   - AI service host.
   - Web frontend host.
   - Secret store for database URL, AI service URL, Entra values, and provider credentials.

4. Add migration and smoke runbooks.
   - Migration up command.
   - Rollback plan.
   - `/healthz`, `/readyz`, `/api/v1/config` checks.
   - Browser path for host creation, player join, card assignment, call, mark, claim, winner, summary, and SSE reconnect.

5. Lock deployment env values.
   - Backend: `APP_ENV=staging`, `DATABASE_URL`, `CORS_ALLOWED_ORIGINS`, `AUTH_MODE`, player timeout settings, AI service config, and telemetry config.
   - Web: `NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_APP_URL`.
   - AI service: provider mode, CORS origins, service token, speech/storage config if enabled.

## Seven-Person Starting Split

| Lane | Owner Focus | First Deliverables |
|---|---|---|
| Release captain | branch hygiene, issue board, merge order | clean `main`, release checklist, deploy scope |
| Platform/cloud | hosting, database, images, secrets | staging resources and backend Dockerfile |
| Backend/auth | Entra verifier, ownership rules, migrations | real JWT verification and auth tests |
| Frontend/product | app env, auth-aware API calls, production flow | auth/session shell and E2E happy path |
| AI/integrations | AI service runtime, Speech/storage, delivery | production server choice and provider config |
| QA/reliability | CI, smoke, load, browser checks | required checks and 50-player load run |
| Security/docs | CORS, headers, rate limits, runbooks | security checklist and deployment runbook |

## First Acceptance Bar

The first staging deployment is credible when:

- CI is required on pull requests.
- `apps/web` builds without presentation/mock mode.
- `backend-go` can migrate and serve against managed Postgres.
- The web app can complete one backend-backed game path.
- Logs and request IDs make failures debuggable.
- A rollback path is documented before the first live pilot.
