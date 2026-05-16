# AI Service Integration Plan

Last updated: 2026-05-16

This document defines how the Go backend should connect to the Python AI service for topic generation, word generation, AI caller narration, Azure Speech audio, and later theme generation.

The central rule is simple: Python owns AI work; Go owns product orchestration and game truth.

## Current Go Implementation Status

Batch 1 and batch 2 of the Go-side integration are implemented for content prep/review/lock, locked decks, caller asset metadata, mock delivery, lobby open/profile state, and structured theme approval.

- Go config now includes `AI_SERVICE_ENABLED`, `AI_SERVICE_BASE_URL`, and `AI_SERVICE_TIMEOUT_SECONDS`.
- When enabled, Go calls Python `POST /ai/v1/game-prep` with the service-to-service request shape below.
- When disabled, Go uses a deterministic local disabled client so local tests and smoke flows do not need Python, Azure OpenAI, or Azure Speech.
- Go validates AI words by trimming whitespace, rejecting blanks, deduping case-insensitively, and requiring at least 24 unique words.
- Go stores generation jobs and review content in `content_generation_jobs`, `generated_game_content`, and `game_run_content_reviews`.
- Host/admin review endpoints are implemented:
  - `POST /api/v1/games/{gameID}/content/prepare`
  - `GET /api/v1/games/{gameID}/content`
  - `PATCH /api/v1/games/{gameID}/content`
  - `POST /api/v1/games/{gameID}/content/lock`
- Lock creates an approved `ai_generated` word set from the final words and attaches it to the game run so existing card assignment can use the locked content.
- Lock also creates a deterministic `game_call_deck` from a stored seed and shuffle version. Live calls follow this deck when it exists, and each called word links back to the deck item.
- Go client methods now cover Python `POST /ai/v1/caller-assets/bulk`, `POST /ai/v1/caller-assets`, and `POST /ai/v1/themes/generate`, with deterministic local disabled behavior for tests and smoke flows.
- `caller_assets` stores caller lines plus optional audio URL/storage key/status per deck item. Failed or missing AI rows keep fallback text so live gameplay is not blocked by Azure Speech.
- `delivery_batches` and `delivery_attempts` store local/mock player invite attempts with `/join?code={CODE}` links. Real Microsoft Graph, Teams, and email delivery are still deferred.
- `theme_generation_jobs`, `themes`, and `theme_approvals` store structured theme drafts, host edits, approval/rejection, and game application through settings.
- Manual host/admin hooks now exist for content prepare, content lock, caller asset generation, player invites, and lobby open.

Still deferred: frontend wiring, lobby chat, real Microsoft Graph/Teams/email delivery, Azure Speech execution/storage integration, real Azure deployment, rewards, voice claim recordings, and voice-profile consent flows.

## Ownership Boundary

### Go Backend Owns

- Game templates and scheduled game runs.
- Host, player, and admin permissions.
- Player allowlists and game codes.
- Review windows, lock windows, lobby open time, and start time.
- Approved word lists and the final randomized call deck.
- Card assignment, called words, marks, claims, winners, summaries, and audit history.
- Emails, Teams delivery commands, and player invite links.
- Durable persistence in Postgres.
- Realtime events to the frontend through the game event outbox and SSE.

### Python AI Service Owns

- Weekly topic generation.
- Word list generation from the weekly topic and host settings.
- Caller sentence generation.
- Azure Speech text-to-speech audio generation.
- Future AI-generated visual theme drafts.
- Future AI-generated recap copy, if needed.

Python should never decide the official next word, card state, valid claims, winners, player access, game status, rewards, or audit outcomes.

## Scheduled Game Preparation Flow

The full-scale scheduled game flow should work like this.

### T Minus 60 Minutes: Prepare Review Package

Go starts a preparation job for the scheduled game run.

Go sends Python a generation request that includes:

- Game run ID.
- Host settings.
- Requested topic or topic constraints.
- Audience/context metadata that is safe to share with AI.
- Word count target.
- Difficulty/tone/style settings.
- Any excluded words or approved reusable word set.
- Caller and theme preferences, if already configured.

Python returns:

- Weekly topic.
- Generated word list.
- Short host-facing summary.
- Optional caller tone/style suggestion.
- Optional theme suggestion.

Go stores the generated draft content and sends the host an email summary with:

- Week's topic.
- Generated word list.
- Game time and lobby time.
- Game code.
- Player count and allowlist summary.
- Voice/caller settings.
- Other important run details.
- A review/edit link.

The host should see the word list, topic settings, and summary details. The host should not need to see the final randomized call order.

### T Minus 60 To T Minus 30: Optional Host Review

Host review is optional.

- If the host does nothing, the game continues automatically.
- If the host edits the topic, words, caller settings, voice, or other pre-game details inside this window, Go stores the edited version.
- The edited version becomes the input for lock and voice generation.

Host review exists for a quick glance and correction, not as a mandatory approval gate.

### T Minus 30: Lock Game Content

Go locks the game content.

At lock time, Go should:

- Freeze the final approved word list.
- Randomize the call deck.
- Store either the ordered call deck or a shuffle seed plus deterministic shuffle version.
- Freeze caller settings, voice settings, theme settings, and gameplay settings.
- Prevent late edits that would invalidate generated audio.
- Write audit events for content lock and any host edits.

Randomization still happens, but it happens before audio generation so sequence-specific caller lines can be generated safely.

### T Minus 30 To T Minus 10: Bulk Generate Caller Audio

After lock, Go asks Python to generate caller lines and Azure Speech audio for the locked ordered deck.

Python should support a bulk endpoint and return per-word results:

- Called word ID or call deck item ID.
- Sequence number.
- Word.
- Caller sentence.
- Audio asset URL or storage key.
- Generation status.
- Fallback text if generation fails.
- Provider metadata.

Go stores those results against the locked call deck. Go may retry failed items, but failed AI generation must not block the game from calling words. If an item is missing audio at play time, Go can use fallback text or request single-call regeneration.

### Player Invite Delivery

After content lock, Go sends player invites through email first and Teams later.

Invite links can include the public game code, for example:

```text
https://app.example.com/join?code=ABCD12
```

Players must still sign in. The code fills in the game lookup, but access is enforced by Go against the authenticated user's allowlist entry. If players open the Teams tab or visit the site normally, they can type the code manually or sign in and see accessible running/upcoming games.

### T Minus 10: Lobby Opens

The lobby opens 10 minutes before game start.

Production lobby features:

- Player join/rejoin.
- Access check against the allowlist.
- Connected player list/count.
- Icon/avatar selection.
- Game details.
- Optional chat later.

Chat should be planned as a later feature so it does not slow down the core gameplay path.

### Game Start And Live Play

At start time, Go starts the game if the configured start rules are satisfied, such as a minimum player count.

During live play:

- Go calls words from the locked randomized deck.
- Go broadcasts `word.called` events.
- Go includes or exposes the pre-generated caller line/audio for the called word.
- The frontend plays already-generated audio with no Azure Speech wait on the live path.
- Players mark cards or use assisted/auto-mark modes.
- Go validates claims deterministically.
- Go records winners and summary state.

Python is only used live for fallback or regeneration. The normal live path should not wait on AI.

## AI Service API Shape

These are service-to-service contracts between Go and Python. They are not the public frontend API.

### Generate Review Package

```http
POST /ai/v1/game-prep
```

Request:

```json
{
  "gameRunId": "uuid",
  "topicPrompt": "workplace collaboration",
  "wordCount": 75,
  "tone": "fun",
  "audience": "internal workplace team",
  "excludedWords": [],
  "settings": {
    "callerStyle": "light workplace humor",
    "themeMode": "ai_generated"
  }
}
```

Response:

```json
{
  "topic": "Collaboration Week",
  "summary": "A light workplace-themed bingo game focused on team habits and project moments.",
  "words": ["Synergy", "Standup", "Roadmap"],
  "callerStyle": "light workplace humor",
  "themePrompt": "modern office game night"
}
```

### Bulk Generate Caller Assets

```http
POST /ai/v1/caller-assets/bulk
```

Request:

```json
{
  "gameRunId": "uuid",
  "voiceName": "en-US-JennyNeural",
  "tone": "fun",
  "deck": [
    {
      "callDeckItemId": "uuid",
      "word": "Synergy",
      "sequence": 1
    }
  ]
}
```

Response:

```json
{
  "gameRunId": "uuid",
  "assets": [
    {
      "callDeckItemId": "uuid",
      "word": "Synergy",
      "sequence": 1,
      "line": "Our first word is Synergy. Somehow, it already scheduled a follow-up meeting.",
      "audioUrl": "https://storage.example.com/audio/synergy.mp3",
      "status": "ready"
    }
  ]
}
```

### Single Caller Asset Fallback

```http
POST /ai/v1/caller-assets
```

Use this only when one item failed, a host regenerates one line before lock, or the live game needs a fallback.

### Theme Generation

```http
POST /ai/v1/themes/generate
```

Request:

```json
{
  "gameRunId": "uuid",
  "prompt": "Christmas",
  "tone": "festive",
  "allowedAssets": ["sparkles", "briefcase", "rocket", "trophy", "coffee", "star", "confetti", "subtle", "none"]
}
```

Response:

```json
{
  "name": "Christmas",
  "summary": "A festive but readable holiday theme for a workplace bingo game.",
  "palette": {
    "primary": "#0F7B3F",
    "accent": "#C62828",
    "background": "#F7FBF8"
  },
  "icons": ["star", "sparkles"],
  "decorations": ["confetti"],
  "motion": "subtle",
  "callerTone": "festive",
  "accessibility": {
    "contrastChecked": true,
    "avoidColorOnlyMeaning": true
  }
}
```

## Implemented Go Data Concepts

- `content_generation_jobs`
- `generated_game_content`
- `game_run_content_reviews`
- `game_call_deck`
- `caller_assets`
- `theme_generation_jobs`
- `themes`
- `theme_approvals`
- `delivery_batches`
- `delivery_attempts`

`game_call_deck` preserves:

- Game run ID.
- Word text and optional word set word ID.
- Sequence.
- Randomization seed/version.
- Locked timestamp.
- Linked called word ID when called.

`caller_assets` preserves:

- Game run ID.
- Call deck item ID.
- Word.
- Caller sentence.
- Audio URL or blob key.
- Voice name.
- Provider.
- Status: `pending`, `ready`, `failed`, `fallback`.
- Error reason, if any.

## Theme Generator Plan

The theme generator should follow the same AI boundary as caller content.

Python generates a safe draft theme. Go stores it, controls approval, and exposes the approved theme through snapshots/settings for the future frontend.

The AI service should return structured tokens, not executable code.

Example:

```json
{
  "name": "Christmas",
  "summary": "A festive but readable holiday theme for a workplace bingo game.",
  "palette": {
    "primary": "#0F7B3F",
    "accent": "#C62828",
    "background": "#F7FBF8"
  },
  "icons": ["snowflake", "gift", "tree"],
  "decorations": ["subtle_snow"],
  "motion": "subtle",
  "callerTone": "festive",
  "accessibility": {
    "contrastChecked": true,
    "avoidColorOnlyMeaning": true
  }
}
```

Rules:

- No arbitrary CSS.
- No JavaScript.
- No unapproved external image URLs.
- No unsafe asset references.
- Frontend maps tokens to a controlled set of visual options.
- Host can preview/edit/approve before the theme applies.
- Theme can influence caller tone, but it cannot change game rules.

## Lobby Chat Plan

Lobby chat should be a later feature after backend gameplay and frontend wiring are stable.

Recommended implementation:

- Go owns chat messages and moderation state.
- Store chat messages in Postgres for the game run.
- Broadcast `chat.message_created` through the existing game event outbox/SSE path.
- Add rate limiting and basic moderation before enabling broadly.
- Keep chat disabled after game finish unless summary/reactions are explicitly wanted.
- Python can optionally provide moderation suggestions later, but Go should enforce final moderation rules.

Likely APIs:

- `GET /api/v1/games/{gameID}/chat/messages`
- `POST /api/v1/games/{gameID}/chat/messages`
- `DELETE /api/v1/games/{gameID}/chat/messages/{messageID}`
- `POST /api/v1/games/{gameID}/chat/messages/{messageID}/report`

Chat is useful for lobby energy, but it should not be on the critical path for the first production gameplay integration.

## Implementation Priority

1. Add game prep/content generation job concepts.
2. Add host review email and optional edit window.
3. Add content lock at T minus 30.
4. Add locked randomized call deck storage.
5. Add Python bulk caller asset generation.
6. Store caller assets and expose them through snapshots/events.
7. Send player invites after lock.
8. Open lobby at T minus 10 with icon/avatar selection.
9. Add theme generation as a structured draft/approval feature.
10. Add lobby chat after the core game loop is stable.
