# Documentation Index

## Read First

- `proposals/TECH_STACK_PROPOSAL_UPDATED.md` - current recommended stack.
- `architecture/AUTONOMOUS_BACKEND_ARCHITECTURE.md` - product and backend architecture decisions.
- `architecture/GO_BACKEND_PLAN.md` - lower-level Go backend implementation plan.
- `architecture/V1_IMPLEMENTATION_PLAN.md` - production V1 build order and next realtime/game-engine milestones.
- `architecture/FULL_SCALE_DEPLOYMENT_ROADMAP.md` - full CGI/Azure/Microsoft 365/AI/rewards deployment roadmap beyond V1.
- `architecture/V1_REALTIME_BACKBONE_PROMPT.md` - ready-to-use prompt for the next realtime backend implementation pass.

## Supporting Docs

- `architecture/PRODUCTION_READINESS_RESEARCH.md` - earlier production-readiness gap analysis.
- `product/Virtual_Bingo_BRD_Updated.docx` - source BRD for the current manual process.
- `proposals/original/Virtual Bingo Technical Stack Proposal.pdf` - original stack proposal.
- `ui-demo/design.md` and `ui-demo/system_design.md` - UI/demo specifications.
- `ui-demo/FRONTEND_DEMO_POLISH_REVIEW.md` - notes on the demo app.
- `../backend-go/README.md` - local Go backend run/test commands.

## Notes

The current `apps/frontend-demo/` app is kept as evidence and a demo reference. Future production work should be planned around the Azure + Go + Python AI service architecture in the proposal and architecture docs.

For local development, `../docker-compose.yml` currently provides Postgres for the Go backend foundation.
