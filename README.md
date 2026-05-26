# Virtual Bingo Workspace

This workspace contains the Virtual Bingo web app, Go API, AI service, deployment notes, and product reference docs.

## Folder Map

- `backend-go/` - production Go API foundation for the autonomous platform.
- `apps/web/` - Next.js web app.
- `ai-service/` - Python AI/content and caller-asset service.
- `docs/architecture/` - current production architecture and backend planning.
- `docs/proposals/` - updated technical stack proposal and original proposal PDF.
- `docs/product/` - source BRD from the current manual process.
- `docker-compose.yml` - local development services, starting with Postgres.

## Current Source Of Truth

Start with:

1. `docs/proposals/TECH_STACK_PROPOSAL_UPDATED.md`
2. `docs/architecture/AUTONOMOUS_BACKEND_ARCHITECTURE.md`
3. `docs/architecture/GO_BACKEND_PLAN.md`
4. `docs/architecture/V1_IMPLEMENTATION_PLAN.md`
5. `docs/architecture/DEPLOYMENT_READINESS_AUDIT.md`

## Local Backend Quick Start

```bash
docker compose up -d postgres
cd backend-go
cp .env.example .env
DATABASE_URL=postgres://bingo:bingo@localhost:5432/virtual_bingo?sslmode=disable go run ./cmd/migrate up
go run ./cmd/api
```

Then check:

- `http://localhost:8080/healthz`
- `http://localhost:8080/readyz`
- `http://localhost:8080/api/v1/version`

## Local Web Quick Start

```bash
cd apps/web
cp .env.example .env
npm install
npm run dev
```
