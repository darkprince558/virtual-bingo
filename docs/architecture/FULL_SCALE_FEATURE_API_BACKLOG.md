# Full-Scale Feature API Backlog

Last updated: 2026-05-16

This document captures the full-scale product features that extend beyond the local live-game backend. These features should stay behind explicit API contracts and host/admin settings so the game remains reliable while still feeling fun.

The core rule still holds: the Go backend owns game truth. Teams, microphone recordings, AI caller scripts, theme generation, and auto-marking should enhance the experience without replacing deterministic card assignment, word calls, claim validation, winner ordering, audit history, or permissions.

## Feature Decisions

### Teams App Experience

Players should eventually be able to access the game through Microsoft Teams.

Preferred shape:

- Package the web app as a Teams app/tab so players can play inside Teams without building a second game client.
- Use Teams bot/proactive messages for invites, reminders, re-entry links, winner announcements, and summaries.
- Keep the same backend API and game engine whether the user opens the game in Teams or in a browser.
- Treat Teams as an access and delivery surface, not the source of game state.

### Voice Bingo Claim Recording

Hosts can optionally require or allow players to say "Bingo" into the microphone when claiming.

Preferred claim flow:

1. Player presses `Claim Bingo`.
2. Frontend records a short clip, ideally 2-4 seconds.
3. Speech recognition verifies that the player said "Bingo".
4. Frontend uploads the short recording.
5. Backend creates the claim and attaches the recording metadata.
6. Backend validates the claim deterministically against persisted card/call state.
7. Realtime event notifies host/players that a claim happened.
8. If configured, every game client plays the submitted "Bingo" recording as a shared live moment.

Important boundaries:

- The recording proves the player performed the claim action; it does not prove the card is valid.
- The server-side bingo validation remains authoritative.
- Clips should be short, claim-scoped, and not reused as voice profiles.
- Host settings should control whether voice claim is off, optional, or required.
- Host settings should also control whether recordings autoplay for everyone.
- Accessibility/privacy fallback should exist when microphone permission fails.

### Auto-Mark Mode

Hosts can allow or force assisted card marking.

Recommended modes:

- `manual`: player marks cells manually.
- `assist`: matching cells are highlighted after a word is called, but the player confirms the mark.
- `auto_mark`: matching cells are marked automatically when words are called.

Important boundaries:

- Auto-marking changes card marks, not the called word sequence.
- Auto-marking should not auto-submit a claim by default.
- The UI can show "Ready to claim" when a valid pattern appears, but the player should still claim.
- Host/template settings decide whether players can choose their own marking mode.

### AI Caller Sentences

The AI caller should say full sentences that include the called bingo word.

Example:

- Called word: `Synergy`
- Caller sentence: `Our next word is Synergy, the secret ingredient in every meeting that somehow creates three follow-up meetings.`

Important boundaries:

- The Go backend commits the called word first.
- AI generates text and audio for that exact word after the word is committed.
- The sentence must include the exact called word clearly.
- The UI must still display the called word visually.
- If AI generation or audio fails, fallback text/audio should be simple: `Next word is Synergy.`
- Caller output is experience flavor, not game truth.

### Theme Generator

Hosts can type a theme prompt such as "Christmas" and generate a themed visual layer for the game.

Preferred shape:

- AI returns a structured theme object, not arbitrary CSS or executable code.
- The frontend maps safe theme tokens to colors, icons, decorations, motion, and optional caller tone.
- Themes should preserve accessibility, contrast, and gameplay readability.
- The host should preview and approve the theme before it applies to a live game.

Example structured output:

```json
{
  "name": "Christmas",
  "palette": {
    "primary": "#0F7B3F",
    "accent": "#C62828",
    "background": "#F7FBF8"
  },
  "icons": ["snowflake", "gift", "tree"],
  "decorations": ["subtle_snow"],
  "tone": "festive"
}
```

Important boundaries:

- Do not let AI generate arbitrary CSS, JavaScript, external image URLs, or unsafe asset references.
- Store the generated theme object and host approval state.
- Apply themes at template/run level so recurring games can reuse them.

## API Surface Needed

### Teams App And Delivery APIs

These are needed when the Teams app/tab and Teams bot are built:

- `GET /api/v1/teams/config`
  - Returns Teams app metadata needed by the frontend, such as app ID, tenant constraints, and enabled Teams capabilities.
- `POST /api/v1/teams/installations`
  - Records or refreshes a Teams app installation context for a user/team/chat.
- `GET /api/v1/teams/context`
  - Resolves a Teams tab context to the current user, tenant, team/channel/chat, and game/template if present.
- `POST /api/v1/games/{gameID}/teams/invites`
  - Sends Teams invites or posts game links through the configured Teams delivery target.
- `POST /api/v1/games/{gameID}/teams/reminders`
  - Sends reminder messages.
- `POST /api/v1/games/{gameID}/teams/summary`
  - Sends winner/summary messages.
- `GET /api/v1/games/{gameID}/deliveries`
  - Implemented for local/mock delivery attempts; later it should also list real Teams and email attempts.
- `POST /api/v1/deliveries/{deliveryID}/retry`
  - Implemented as a local/mock retry-state reset; later it should call the real provider retry path.
- `POST /api/v1/games/{gameID}/deliveries/player-invites`
  - Implemented for local/mock email-style invite attempts with `/join?code={CODE}` links from the allowlist.

Likely data concepts:

- `teams_installations`
- `teams_conversation_refs`
- `delivery_batches`
- `delivery_attempts`

Status: the generic delivery log/retry foundation exists. Real Teams, Graph, and email provider delivery remains deferred.

### Voice Claim APIs

These support microphone "Bingo" claims and shared playback:

- `POST /api/v1/games/{gameID}/claim-recordings`
  - Creates an upload request or accepts a small direct upload for a short claim recording.
  - Response includes recording ID, upload status, duration limit, and playback metadata.
- `POST /api/v1/games/{gameID}/claim-recordings/{recordingID}/verify`
  - Stores speech-to-text result, confidence, and whether the word "Bingo" was detected.
- `POST /api/v1/games/{gameID}/claims`
  - Extend existing claim submit API to optionally accept `claimRecordingId`.
- `GET /api/v1/games/{gameID}/claims/{claimID}/recording`
  - Returns signed/authorized playback metadata for the attached recording.
- `POST /api/v1/games/{gameID}/claims/{claimID}/recording/playback-events`
  - Optional analytics/audit endpoint for whether clients attempted playback.

Realtime events:

- `claim.recording_uploaded`
- `claim.recording_verified`
- `claim.submitted`
- `claim.recording_ready`
- `winner.created`

Likely data concepts:

- `claim_recordings`
- `claim_recording_transcripts`
- `game_run_voice_claim_settings`

Settings needed on game/template:

- `voice_claim_mode`: `off`, `optional`, `required`
- `voice_claim_autoplay`: boolean
- `voice_claim_play_scope`: `host_only`, `all_players`, `winner_only`
- `voice_claim_max_seconds`

### Auto-Mark APIs

Auto-mark now has a backend foundation through durable game settings, player preferences, transactional called-word marking, a backfill endpoint, snapshots, and claim-readiness. Template defaults remain future work.

- `GET /api/v1/games/{gameID}/settings`
  - Implemented for host/admin reads. Lazily creates default `game_run_settings`.
- `PATCH /api/v1/games/{gameID}/settings`
  - Implemented for host-controlled game-run settings such as mark mode, player override permission, claim-readiness visibility, and placeholder voice/caller/theme fields.
- `PATCH /api/v1/templates/{templateID}/settings`
  - Planned. Stores default settings for recurring games after template APIs are active.
- `GET /api/v1/games/{gameID}/players/me/preferences`
  - Implemented for current-player marking preference reads.
- `PATCH /api/v1/games/{gameID}/players/me/preferences`
  - Implemented for player-selected marking mode when host allows player choice.
- `POST /api/v1/games/{gameID}/auto-mark/run`
  - Implemented as host/admin idempotent backfill after reconnects or mode changes.
- `GET /api/v1/games/{gameID}/players/me/claim-readiness`
  - Implemented as a current-player UX helper from persisted card/called-word state. This does not replace authoritative claim validation.

Realtime events:

- `card.auto_marked` - implemented with compact affected counts.
- `game.settings_updated` - implemented.
- `player.preferences_updated` - implemented.
- `player.claim_ready` - planned only if the product needs a proactive event later; current readiness is pull-based.

Implemented data concepts:

- `game_run_settings`
- `player_preferences`

Settings implemented:

- `marking_mode`: `manual`, `assist`, `auto_mark`
- `allow_player_marking_mode_choice`: boolean
- `show_claim_readiness`: boolean
- placeholder settings: `voice_claim_mode`, `voice_claim_autoplay`, `caller_mode`, `theme_mode`, `theme_id`

Behavior implemented:

- `manual` leaves existing player mark/unmark behavior unchanged.
- `assist` does not mark cells automatically, but claim-readiness still works as a UX helper.
- `auto_mark` marks matching unmarked cells when words are called, using case-insensitive whitespace-trimmed matching.
- Auto-mark does not submit claims and does not create winners outside the existing claim flow.
- The backfill endpoint is idempotent and reports `playersScanned`, `playersMarked`, `calledWordsScanned`, `cellsMarked`, `mode`, and `skippedReason` when nothing applies.

### AI Caller Sentence And Audio APIs

These are partly implemented for the post-lock, pre-game path. Go now owns `game_call_deck`, creates a deterministic locked deck at content lock, calls through that deck from `POST /api/v1/games/{gameID}/calls`, and stores deck-linked `caller_assets` rows from Python bulk generation or local disabled fallback text.

Implemented:

- `POST /api/v1/games/{gameID}/caller-assets/generate`
  - Generates/stores caller metadata for the full locked deck through Python `POST /ai/v1/caller-assets/bulk` or the disabled local AI client.
- `POST /api/v1/games/{gameID}/calls`
  - Calls the next locked deck item when a deck exists and includes ready caller metadata in the response.
- Host/player snapshots expose current caller metadata.

Still future for richer live caller control:

- `POST /api/v1/games/{gameID}/caller/next`
  - Host/worker command that calls the next word and starts caller generation for that committed word.
  - Could wrap existing `POST /api/v1/games/{gameID}/calls` later.
- `POST /api/v1/games/{gameID}/calls/{calledWordID}/caller-line`
  - Generates or regenerates a sentence for the called word.
- `GET /api/v1/games/{gameID}/calls/{calledWordID}/caller-line`
  - Returns text, status, safety result, and fallback state.
- `POST /api/v1/games/{gameID}/calls/{calledWordID}/audio`
  - Generates TTS audio from approved caller text.
- `GET /api/v1/games/{gameID}/calls/{calledWordID}/audio`
  - Returns authorized playback metadata.
- `POST /api/v1/games/{gameID}/caller/pause`
- `POST /api/v1/games/{gameID}/caller/resume`
- `POST /api/v1/games/{gameID}/caller/skip`
- `POST /api/v1/games/{gameID}/caller/replay`
  - Host controls for the live caller.
- `PATCH /api/v1/games/{gameID}/caller/settings`
  - Voice, cadence, sentence style, fallback behavior.

Realtime events:

- `caller.line_generating`
- `caller.line_ready`
- `caller.audio_generating`
- `caller.audio_ready`
- `caller.playback_started`
- `caller.playback_finished`
- `caller.failed`

Likely data concepts:

- `game_call_deck` - implemented.
- `caller_assets` - implemented for line/audio metadata rows and fallback status.
- `caller_settings`
- `caller_lines`
- `caller_audio_assets`
- `caller_playback_events`

Settings needed:

- `caller_mode`: `off`, `text_only`, `tts`
- `caller_voice_id`
- `caller_style`
- `call_cadence_seconds`
- `require_word_in_sentence`: true

### Theme Generator APIs

These support AI-generated themed UI without unsafe arbitrary code. The game-run theme APIs are implemented; template-level theme application remains future.

- `POST /api/v1/themes/generate`
  - Implemented. Input: theme prompt, optional tone, accessibility constraints, game context.
  - Output: draft structured theme object.
- `GET /api/v1/themes/{themeID}`
  - Implemented.
- `PATCH /api/v1/themes/{themeID}`
  - Implemented for draft safe theme fields.
- `POST /api/v1/themes/{themeID}/approve`
  - Implemented for host/admin approval.
- `POST /api/v1/themes/{themeID}/reject`
  - Implemented.
- `POST /api/v1/games/{gameID}/theme`
  - Implemented.
- `POST /api/v1/templates/{templateID}/theme`
  - Future: applies an approved theme to a recurring template.
- `GET /api/v1/theme-assets`
  - Implemented. Lists safe built-in icons/decorations/motion presets the AI is allowed to reference.

Realtime events:

- `theme.generated`
- `theme.approved`
- `theme.applied`

Likely data concepts:

- `themes`
- `theme_generation_jobs`
- `theme_approvals`
- `theme_asset_catalog`

Status: `themes`, `theme_generation_jobs`, and `theme_approvals` are implemented. A database-backed `theme_asset_catalog` is still future; current safe assets are code-defined.

Settings needed:

- `theme_mode`: `default`, `manual`, `ai_generated`
- `allow_live_theme_changes`: boolean
- `theme_id`

## Settings Foundation Follow-On Plan

The current `game_run_settings` table is the connection point for later full-scale features, but those features should remain feature-flagged off until their real auth, storage, privacy, and provider flows exist.

### Voice Bingo Claim

Future voice claims should use the existing `voice_claim_mode` and `voice_claim_autoplay` fields:

- `voice_claim_mode=off`: current behavior, no recording required.
- `voice_claim_mode=optional`: player can attach a short "Bingo" recording to a claim.
- `voice_claim_mode=required`: claim submission requires a verified recording path.
- `voice_claim_autoplay=true`: clients may play approved claim audio after server authorization.

Future tables:

- `claim_recordings`
- `claim_recording_transcripts`

Future endpoints:

- Create/upload a short recording or signed upload request.
- Verify speech-to-text result and "Bingo" detection.
- Attach `claimRecordingId` to existing claim submission.
- Read authorized playback metadata.

Do not implement microphone upload, Azure Speech, playback, or transcript storage until privacy and retention rules are explicit.

### AI Caller

AI caller work uses `caller_mode`:

- `off`: current word calls only.
- `text_only`: generate caller lines for committed words.
- `tts`: generate caller lines plus audio assets after text is approved or safely generated.

Implemented/future tables:

- `game_call_deck` - implemented.
- `caller_assets` - implemented.
- `caller_settings` - future richer controls.
- `caller_lines` - represented for now by `caller_assets.line`.
- `caller_audio_assets` - represented for now by `caller_assets.audio_url/storage_key`; real storage integration is future.

Future endpoints:

- Generate/regenerate caller text for a committed `calledWordID`.
- Generate TTS audio for approved caller text.
- Read caller line/audio status and playback metadata.
- Host controls for pause, resume, replay, skip, and cadence.

The called word must always be committed first, and generated text/audio must be flavor only.

### Theme Generator

Theme generation uses `theme_mode` and `theme_id`:

- `default`: current frontend theme behavior.
- `manual`: host-selected safe theme tokens.
- `ai_generated`: host-approved AI-generated structured theme.

Implemented tables:

- `themes`
- `theme_generation_jobs`
- `theme_approvals`

Implemented/future endpoints:

- Generate a structured theme from a prompt - implemented for game context.
- Read/edit safe theme fields - implemented.
- Approve/reject a generated theme - implemented.
- Apply an approved theme to a game/template - implemented for games; templates future.

AI must return structured tokens only, not arbitrary CSS, JavaScript, external image URLs, or executable assets.

### Teams App

Teams APIs should wait until real Entra auth and Microsoft app registration are ready. The future shape is:

- Teams tab/context APIs for resolving tenant/team/channel/chat context.
- Delivery APIs for invites, reminders, re-entry links, winner announcements, and summaries.
- Installation/context storage for Teams app and bot delivery.

Teams should remain an access and delivery surface. The Go backend and Postgres stay the source of game truth.

## API Build Priority

These features should not block the immediate local management APIs. A sensible order is:

1. Done: add game-run settings primitives and player marking preferences.
2. Done: add auto-mark settings, transactional auto-mark, backfill, and claim-readiness support.
3. Add template-level defaults for settings after recurring template APIs are active.
4. Add voice claim recording metadata and claim attachment support after privacy/storage rules are decided.
5. Done: add locked call deck plus caller sentence/audio metadata APIs with fallback behavior. Rich playback controls and real Azure Speech storage remain future.
6. Add Teams tab/context APIs after real Entra auth and Microsoft app registration are working.
7. Add Teams bot delivery APIs after delivery logs and retry tracking exist. The local/mock delivery log foundation is now done.
8. Done: add theme generator APIs and safe token persistence. Frontend rendering remains future.

## Open Questions

- Should a voice recording autoplay for everyone on every submitted claim, or only valid/confirmed claims?
- Should hosts be allowed to replay a player's "Bingo" recording?
- Should auto-mark be a host-forced setting, a player preference, or both?
- Should AI caller sentences be generated live per call, pre-generated before the game, or both with live fallback?
- Should generated themes affect only visuals, or also caller tone and generated word sets?
- Should Teams gameplay be a tab-only experience first, with bot commands later?
