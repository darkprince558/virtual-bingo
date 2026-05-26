# Virtual Bingo Web App

This is the Next.js web app for Virtual Bingo. It talks to the Go API through `NEXT_PUBLIC_API_URL` and should be deployed with real backend URLs, not local development fallbacks.

The main host/player paths are wired to the Go backend where those routes exist. Admin requests, template management, voice settings, and rewards still need production backend work before an employee pilot.

## Run Locally

Start the Go backend first. The web app expects `NEXT_PUBLIC_API_URL` to include the `/api/v1` prefix:

```bash
NEXT_PUBLIC_API_URL="http://localhost:8080/api/v1"
```

Then run the app:

```bash
npm install
npm run dev
```

## Build

```bash
npm run lint
npm run build
```

## Notes

- Landing, host setup, AI content review, host live control, lobby, player game, activity feed, claim acknowledgements, and summary use backend-backed flows where available.
- The AI review screen uses the Go content pipeline: prepare, edit, lock, generate caller assets, and open lobby. With `AI_SERVICE_ENABLED=false`, the backend uses deterministic local fallback content.
- The host setup panel on `/host` can edit game name/code/winning pattern, choose a word set, update marking/readiness settings, and bulk add allowed players.
- Player game sends heartbeat calls while open, shows reconnect missed-word notices, and displays claim-readiness before submission. Backend claim validation remains authoritative.
- `/summary?gameId={id}` reads the backend summary and shows empty states when no winners exist yet.
- Local development auth is header-based and must not be used for a real employee-facing deployment.

