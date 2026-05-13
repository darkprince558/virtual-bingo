# Virtual Bingo Workspace

This workspace contains the planning docs, source BRD/proposal files, and the old frontend demo for the Virtual Bingo project.

## Folder Map

- `backend-go/` - production Go API foundation for the autonomous platform.
- `docs/architecture/` - current production architecture and backend planning.
- `docs/proposals/` - updated technical stack proposal and original proposal PDF.
- `docs/product/` - source BRD from the current manual process.
- `docs/ui-demo/` - UI/demo design notes and polish review.
- `apps/frontend-demo/` - the Next.js demo app built for manager review.
- `archive/generated/` - generated or obsolete root files kept for reference.
- `docker-compose.yml` - local development services, starting with Postgres.

## Current Source Of Truth

Start with:

1. `docs/proposals/TECH_STACK_PROPOSAL_UPDATED.md`
2. `docs/architecture/AUTONOMOUS_BACKEND_ARCHITECTURE.md`
3. `docs/architecture/GO_BACKEND_PLAN.md`
4. `docs/architecture/MVP_IMPLEMENTATION_PLAN.md`

The frontend app is a demo, not the final production architecture.

## Local Backend Quick Start

```bash
docker compose up -d postgres
cd backend-go
cp .env.example .env
go run ./cmd/api
```

Then check:

- `http://localhost:8080/healthz`
- `http://localhost:8080/readyz`
- `http://localhost:8080/api/v1/version`
