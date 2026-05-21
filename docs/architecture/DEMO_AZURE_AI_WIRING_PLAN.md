# Demo Azure And AI Wiring Plan

Last updated: 2026-05-21

This plan is for the May 22, 2026 demo. The goal is a polished, believable product demo without trying to finish the whole Azure production platform overnight.

## Demo Strategy

Use the real Go backend and Postgres for the core game flow. Use the Python AI service for the fun AI surfaces. Keep admin-only, Microsoft 365, Teams, rewards, and deep production infrastructure as mock or narrated roadmap items.

The demo should feel real where the audience can interact:

- Host creates or opens a game.
- Host prepares AI content.
- Host reviews/locks words.
- Host generates caller assets.
- Host opens lobby and runs live calls.
- Player joins, sees a card, marks words, submits a claim, and sees summary state.

Everything around the game can be demo scaffolding:

- Admin approvals can be mock.
- Voice profile management can be mock.
- Rewards can be mock.
- Teams delivery can be mock.
- Real Entra login can be deferred unless it is already available and low-risk.
- Full Azure deployment can be deferred if local demo reliability is higher.

## What To Hook Up For The Demo

### P0: Real Backend Demo Flow

This is the minimum demo spine.

- Run Postgres.
- Run Go backend with migrations and seed/demo game data.
- Run frontend with `NEXT_PUBLIC_API_URL` pointing to the Go API.
- Verify these screens against live backend data:
  - `/host`
  - `/host/review?gameId={gameId}`
  - `/host/live?gameId={gameId}`
  - `/play?gameId={gameId}&playerId={playerId}`
  - `/summary?gameId={gameId}`

Why: this is the part people will actually judge as the game.

### P0: Python AI Service Connected To Go

Turn on the existing service-to-service path:

```env
AI_SERVICE_ENABLED=true
AI_SERVICE_BASE_URL=http://localhost:5001
AI_SERVICE_TIMEOUT_SECONDS=15
```

Run the Python service through Docker Compose or locally. Start in `mock` provider mode first:

```env
AI_PROVIDER_MODE=mock
TTS_PROVIDER=mock
AUDIO_STORAGE_PROVIDER=mock
```

Why: this proves the architecture is real while avoiding Azure credential or network risk during the demo.

### P1: Azure Speech For Caller Audio

Only enable this if the Azure Speech key and region are ready before demo practice.

```env
AI_PROVIDER_MODE=real
TTS_PROVIDER=azure
AUDIO_STORAGE_PROVIDER=mock
AZURE_SPEECH_KEY=...
AZURE_SPEECH_REGION=canadacentral
DEFAULT_VOICE_NAME=en-US-JennyNeural
```

This gives real generated audio bytes, but keeps storage mocked. If playback from persisted URLs is not fully wired in the frontend, present it as "generated caller assets are stored and ready for playback" and show the caller text clearly.

Do not make live word calling wait on Azure Speech. Caller assets should be generated before the live run, and missing audio should fall back to text.

### P2: Azure Blob Storage For Audio URLs

Only enable Blob if there is enough time to test public/container access and playback.

```env
AUDIO_STORAGE_PROVIDER=azure
AZURE_STORAGE_CONNECTION_STRING=...
AZURE_STORAGE_CONTAINER=bingo-narration-audio
```

This is nice, but not demo-critical. A broken Blob permission or private URL will look worse than a clean mocked audio URL plus visible caller lines.

## What Should Stay Mocked Tomorrow

Keep these as mocked or roadmap/demo surfaces:

- Microsoft Entra login.
- Microsoft Graph email delivery.
- Teams app/tab integration.
- Admin host approval workflows.
- Employee voice profiles and consent.
- Voice Bingo claim recordings.
- Rewards/gift cards.
- Redis or Service Bus fanout.
- Full Azure Container Apps deployment.

The honest demo framing is: "Core gameplay and AI prep are running against the real backend contract. Microsoft delivery, auth, rewards, and production cloud scale are the next integration layer."

## Demo Run Order

1. Start the backend demo stack:

   ```bash
   ./scripts/demo-backend.sh
   ```

2. Start the frontend demo app:

   ```bash
   ./scripts/demo-frontend.sh
   ```

3. Open `http://localhost:3004/demo`.
4. Click `Prepare Showcase` to generate/lock content, generate caller assets, and open the lobby.
5. Open `AI Review` and show host control over generated content.
6. Use the player shortcuts to join as Alex, Sam, or Taylor.
7. Open `Live Control` and call words live.
8. Click `Prime Winner Moment` when you need a guaranteed backend-confirmed claim during the presentation.
9. Show host activity and summary.

Optional check after both servers are running:

```bash
./scripts/demo-smoke.sh
```

## Backup Plan

If Azure or Python fails, switch Go back to deterministic fallback:

```env
AI_SERVICE_ENABLED=false
```

The demo still works because Go can generate local placeholder words, caller lines, and safe theme tokens. This is better than debugging cloud credentials in front of people.

If frontend/backend integration fails, use the seeded local demo flow and avoid creating new games live.

If local ports collide, move the frontend to `3003` and set:

```env
NEXT_PUBLIC_API_URL=http://localhost:18081/api/v1
```

## Demo Acceptance Checklist

- `GET /healthz` and `GET /readyz` are healthy.
- `GET /api/v1/config` reports AI content/caller capabilities.
- Frontend host page loads real backend games.
- AI review can prepare content without errors.
- Content lock succeeds.
- Caller asset generation succeeds, with text fallback at minimum.
- Live host screen can call words.
- Player screen receives or refreshes called words.
- Claim flow produces host-visible activity.
- Summary page loads without crashing.

## Recommendation

For tomorrow, prioritize this exact version:

Real Go backend, real Postgres, real frontend wiring, Python AI service in mock mode, optional Azure Speech only after the full local demo path is already reliable.

This gives the demo the right shape: the architecture is credible, the AI path is visible, and the risky external integrations do not get to derail the game.
