# Documentation Index

## Current Source Of Truth

- `proposals/TECH_STACK_PROPOSAL_UPDATED.md` - current recommended stack.
- `architecture/AUTONOMOUS_BACKEND_ARCHITECTURE.md` - product and backend architecture decisions.
- `architecture/GO_BACKEND_PLAN.md` - lower-level Go backend implementation plan.
- `architecture/V1_IMPLEMENTATION_PLAN.md` - production V1 build order and next realtime/game-engine milestones.
- `architecture/FULL_SCALE_DEPLOYMENT_ROADMAP.md` - full CGI/Azure/Microsoft 365/AI/rewards deployment roadmap beyond V1.
- `architecture/DEPLOYMENT_READINESS_AUDIT.md` - current deploy-readiness checklist, verification results, blockers, and seven-person work split.
- `architecture/FULL_SCALE_FEATURE_API_BACKLOG.md` - API backlog for Teams app access, voice Bingo claims, auto-mark mode, AI caller sentences/audio, and AI theme generation.
- `architecture/AI_SERVICE_INTEGRATION_PLAN.md` - Go/Python AI-service boundary, pre-game content lock flow, caller audio generation, theme generation, and later lobby chat plan.
- `../backend-go/README.md` - local Go backend run/test commands.
- `../apps/web/README.md` - local Next.js web app run/build commands.

## Product And Proposal References

- `product/Virtual_Bingo_BRD_Updated.docx` - source BRD for the current manual process.
- `proposals/original/Virtual Bingo Technical Stack Proposal.pdf` - original stack proposal.

## Notes

The current deployable web app lives in `../apps/web/`. Future production work should stay grounded in the Azure + Go + Python AI service architecture in the proposal and architecture docs.

For local development, `../docker-compose.yml` currently provides Postgres for the Go backend foundation.
