# Virtual Bingo Frontend Demo

This is the Next.js wired demo app built for manager review.

The main host/player paths are wired to the Go backend where those routes already exist. Admin, template, voice, and rewards areas are still mock or partial demo surfaces so the full product shape remains visible while the backend is filled in.

It is useful as a runnable product demo, but it is not the final production architecture. The production plan is documented in:

- `../../docs/proposals/TECH_STACK_PROPOSAL_UPDATED.md`
- `../../docs/architecture/AUTONOMOUS_BACKEND_ARCHITECTURE.md`
- `../../docs/architecture/GO_BACKEND_PLAN.md`

## Run Locally

Start the Go backend first. The frontend expects `NEXT_PUBLIC_API_URL` to point at the full API prefix:

```bash
NEXT_PUBLIC_API_URL="http://localhost:18081/api/v1"
```

The legacy `NEXT_PUBLIC_API_BASE_URL` value is still accepted as a fallback, but new local setup should use `NEXT_PUBLIC_API_URL`.

```bash
npm install
npm run dev
```

## Notes

- Landing, host dashboard, setup, AI content review, quick start, host live control, lobby, player game, activity feed, claim acknowledgements, and summary use backend-backed flows where available.
- The AI review screen uses the existing Go content pipeline: prepare, edit, lock, generate caller assets, and open lobby. With `AI_SERVICE_ENABLED=false`, the backend uses local deterministic fallback content; no Python, Azure, or provider credentials are required for this wired demo path.
- The host setup panel on `/host` can edit game name/code/winning pattern, choose a word set, update marking/readiness settings, and bulk add allowed players.
- Player game sends heartbeat calls while open, shows reconnect missed-word notices, and displays claim-readiness before submission. Backend claim validation remains authoritative.
- Host live activity reads `/games/{gameId}/activity`; claim buttons acknowledge backend-confirmed or backend-rejected claims for audit/demo clarity and do not override winner validation.
- `/summary?gameId={id}` reads the real backend summary and shows empty states when no winners exist yet.
- With the local seed, join `LOCAL-DEMO` as `Alex Demo`, `Sam Demo`, or `Taylor Demo` so the generated dev email matches the backend allowed-player list.
- Mock-only or partial surfaces remain for admin requests, templates, voice settings, and rewards.
- Dev auth is header-based and intended for local demo work only.
- The previous AI Studio/Gemini package/env scaffold has been removed.
