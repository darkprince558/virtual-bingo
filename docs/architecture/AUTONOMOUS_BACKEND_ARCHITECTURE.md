# Autonomous Virtual Bingo Backend Architecture

Last updated: 2026-05-12

Related planning docs:

- `../proposals/TECH_STACK_PROPOSAL_UPDATED.md`
- `GO_BACKEND_PLAN.md`
- `../archive/architecture/PRODUCTION_READINESS_RESEARCH.md`

## Goal

The production goal is not just a live bingo website. The goal is an autonomous weekly game operations platform for internal CGI-style workplace bingo.

Managers should set up the game once, choose recurrence and automation preferences, and then the system should handle most of the weekly work:

- Generate or reuse bingo content.
- Create weekly game runs from a recurring schedule.
- Invite the right mailing list / Teams audience.
- Let players join through Microsoft/CGI identity and weekly invite links.
- Run the game with an AI caller that feels live and engaging.
- Validate Bingo automatically.
- Track top 3 winners.
- Email winners and host automatically.
- Trigger gift card fulfillment.
- Keep an admin-visible history of played games.

## Decisions Captured

These are the architecture decisions from the latest product direction.

- This is an internal company game.
- Access should be CGI / Microsoft identity based.
- Microsoft Teams and Outlook integration are first-class requirements.
- Hosts are privileged users.
- Admins can do everything hosts can do, plus view all games and approve host requests.
- Normal users can request host privileges through an approval workflow in the dashboard.
- Realistic game size is around 25 players, but the system should support 100 players.
- Usually one game runs at a time, but multiple concurrent games should be supported.
- Deployment target is Azure.
- The actual game will happen inside the web app for now; Teams and Outlook are invite, reminder, notification, and re-entry channels.
- Players should be able to disconnect and rejoin into the live game from invite/sign-in.
- Game invite links should map to a backend allowlist. If the signed-in CGI/Microsoft user matches the allowed player list for that run, they can join.
- Game patterns should default to random, but hosts can choose allowed patterns manually.
- Card generation should be AI-assisted but editable.
- AI-generated content should require host approval by default, but hosts can opt into fuller autopilot. Hosts should receive pre-game details by email so they can review and fix issues before game time.
- Cards should be assigned fairly like normal bingo, with unique persisted cards per player as the safest default.
- Winner checking should be automatic.
- Winner and host notification should be automatic.
- Players can win multiple times.
- The game tracks top 3 winners.
- Multiple wins by the same player are allowed, and reward values may vary by pattern or placement.
- Managers do not need a deep audit trail, but admins need useful game history and winner/result details.
- On-screen summary and winner-list email are needed first.
- AI caller voice is important and should feel like a live person calling bingo.
- Employee voice profiles will be created from team employee recordings with consent. Admins can add/remove/manage voices. Hosts can select specific approved voices or use automatic voice selection.
- Azure services are the primary AI/cloud provider direction.
- The app should be available to CGI partners broadly, but each game run has host-managed player access.
- Scheduled games should start automatically at the scheduled time only if more than 5 players are present.
- Winning patterns are revealed when the game starts.
- If host approval is required and the host does not review before game time, generated content is auto-approved.
- Admin/security-level immutable logs are required for gift cards, voice profile actions, and AI-generated content.

## Architecture Shape

The backend should become a small Azure-hosted platform with three major parts:

1. Go API service
   - Handles auth, host/admin workflows, sessions, cards, claims, summaries, and frontend APIs.

2. Worker/scheduler service
   - Creates weekly game runs.
   - Kicks off AI card/word generation.
   - Sends invites/reminders.
   - Starts autonomous calling.
   - Sends winner emails and reward fulfillment jobs.

3. Realtime game conductor
   - Drives live game state.
   - Broadcasts current word, AI caller messages, audio state, winners, and leaderboard updates.
   - Can run inside the Go service at first, then split later if needed.

## Recommended Azure Architecture

### Core Runtime

- Frontend: current Next.js app, eventually hosted on Azure Static Web Apps, Azure App Service, or Azure Container Apps.
- Backend API: Go service on Azure Container Apps or Azure App Service.
- Workers: Azure Container Apps Jobs or Azure Functions timer triggers.
- Database: Azure Database for PostgreSQL.
- Queue: Azure Service Bus for async jobs and durable workflows.
- Secrets: Azure Key Vault.
- Realtime: Server-Sent Events or WebSocket from the Go API; Azure SignalR can be considered later if connection scale/platform friction appears.

### Microsoft 365 Integration

- Identity: Microsoft Entra ID.
- Emails: Microsoft Graph `sendMail`.
- Audience access v1: backend allowlist tied to each game invite link.
- Future audience sync: Microsoft Graph groups, configured distribution lists, or Teams team/channel membership.
- Teams delivery: Teams bot proactive messages and/or Teams channel messages.
- Teams meeting creation: deferred. The live game and AI caller should run in-app first.

### AI / Voice

- Word/card generation: Azure OpenAI unless a CGI-mandated Azure-compatible service replaces it.
- Voice calling: Azure AI Speech text-to-speech.
- Replicated employee voices: only through an explicit consent and admin-managed approval flow. Azure personal voice requires explicit user consent with a recorded consent statement, so the product needs a `voice_profiles` approval model, not ad hoc voice cloning.

### Rewards

- Gift card delivery should be implemented as a reward provider adapter.
- The backend should not hard-code one vendor until the actual CGI-approved gift card provider is known.
- The likely direction is emailed voucher-claim links, but the exact provider still needs manager confirmation.
- If gift card codes are pre-purchased, store them encrypted and restrict admin access.
- Prefer provider-side fulfillment when possible so the app triggers delivery rather than storing raw gift card secrets.

## Product Modules

### 1. Identity And Access

Purpose:

- Sign in users through Microsoft/CGI identity.
- Restrict access to internal users.
- Enforce `admin`, `host`, and `player` permissions.

Needed backend concepts:

- Users
- Roles
- Host privilege requests
- Host approvals/rejections
- Admin action history

Important behavior:

- Admins can grant/revoke host privileges.
- Normal users can request host access from the dashboard.
- Hosts can manage their own game templates and game runs.
- Admins can see all game runs and outcomes.
- The app can be available to all CGI partners, but game participation is controlled by each game run's allowlist.

### 2. Recurring Game Templates

Purpose:

- Let a manager configure the weekly recurring game once.

Template should include:

- Name
- Owner/host
- Recurrence rule
- Time zone
- Game duration estimate
- Audience source
- Minimum player threshold
- Teams channel or chat destination
- Outlook/email sender identity
- Card generation mode
- Prompt library settings
- Winning pattern rules
- AI caller settings
- Prize settings
- Manual review requirements

Examples:

- "Run every Friday at 3 PM America/Toronto."
- "Invite members of this Microsoft 365 group."
- "Use a random prompt from this prompt library."
- "Generate 75 words and assign one unique 5x5 card per player."
- "Choose one random winning pattern per game."
- "Start automatically only if more than 5 players have joined."
- "Send gift cards to top 3 winners."

### 3. Game Runs

Purpose:

- Materialize one playable weekly session from a recurring template.

The game run is the concrete event players join.

Game run states:

- `draft`
- `content_generating`
- `content_review`
- `scheduled`
- `invites_sent`
- `lobby_open`
- `live`
- `paused`
- `finished`
- `reward_fulfillment_pending`
- `complete`
- `cancelled`
- `failed`

Important behavior:

- The scheduler creates upcoming runs ahead of time.
- The manager can review generated content if review mode is on.
- The system sends invites before the run.
- Players can rejoin the active run from sign-in or invite link.
- At scheduled start time, the run starts automatically if more than 5 players are present.
- If 5 or fewer players are present, the run should wait, notify the host/admin, or retry according to template settings.

### 4. Audience And Delivery

Purpose:

- Automate weekly invite distribution.

Audience sources:

- Backend-managed allowlist for each game run
- Microsoft 365 group
- Teams team/channel membership
- Configured distribution list
- Manual email list
- Uploaded CSV later if needed

Delivery channels:

- Outlook email
- Teams proactive bot message
- Teams channel post

Needed backend concepts:

- Audience sources
- Audience snapshots
- Invite batches
- Invite deliveries
- Delivery status

Important rule:

- Each game run should snapshot its audience before invites are sent so player eligibility is stable for that run.
- A player can open a shared invite link, but joining requires Microsoft/CGI sign-in and a match against the run allowlist.
- For v1, hosts manage each game/template's allowed-player list.

### 5. AI Content Studio

Purpose:

- Generate fresh bingo word/card content with manager control.

Core concepts:

- Prompt library
- Master prompt list
- Prompt categories
- Prompt selection history
- Generated word sets
- Approved word sets
- Reusable word banks
- Manual edits
- Content generation jobs

Modes:

- Fixed reusable word bank.
- Host enters a prompt for this week.
- System picks a random prompt from the master list.
- System picks a prompt and generates a draft for manager review.
- Fully automatic generation with post-run admin visibility.

Important behavior:

- Default mode should be generated-draft plus host approval.
- Hosts should receive a pre-game review email with game details, generated word set, selected pattern rules, audience count, and review/change link.
- Hosts can enable full autopilot for trusted templates.
- If the host does not review before the configured deadline/game time, the generated content is auto-approved so the recurring game can continue autonomously.
- Generated words should be reviewed for duplicates, banned terms, weak entries, and company appropriateness.
- AI output should be stored as draft content before approval/use.
- Manual edits should preserve who changed what.
- Generated content should be reusable later.

### 6. Card Assignment

Purpose:

- Fairly assign each player a card for that game run.

Recommended rule:

- Generate one unique persisted card per joined player.
- Use a server-side seed for reproducibility.
- Keep the card tied to the player and game run.
- Never let players switch cards after assignment.

Open detail:

- "Normal bingo" can mean several things depending on house rules. For this app, unique persisted cards are the safest and easiest to defend.

### 7. Autonomous Caller

Purpose:

- Replace the manual caller with an engaging AI caller.

Responsibilities:

- Pick next word according to game cadence.
- Generate or select a short spoken line.
- Produce text display for the UI and Teams chat.
- Produce voice audio for playback.
- Broadcast current word and caller line to all players.
- Continue until top 3 winners are found or game ends.

Caller modes:

- Fully automatic.
- Host supervised with pause/skip/next controls.
- Manual fallback if AI or audio fails.

Voice modes:

- Default Azure neural voice.
- Approved internal voice profile with explicit consent.
- Host-selected approved voice.
- Random approved voice per game.
- Random approved voice per word/call, if this is not too chaotic.
- Auto mode that chooses from approved voices using template/game settings.

Important safety/compliance requirement:

- Do not clone or replicate an employee voice unless that person has explicitly consented and the voice profile is approved for this use.
- Admins manage voice creation, approval, deactivation, and removal.
- Hosts can only select from approved active voice profiles.

### 8. Live Game Engine

Purpose:

- Maintain the live state players see.

Responsibilities:

- Session status.
- Current word.
- Called word history.
- Player card mark state.
- Claim state.
- Winner order.
- Leaderboard.
- Connection/rejoin state.
- Realtime broadcasts.
- Scheduled start threshold.
- Pattern reveal at game start.

Recommended v1 protocol:

- REST for mutations.
- SSE for live updates.
- Frontend refetches host/player snapshot after important events.

Why:

- For 25 to 100 players, this is simpler than full bidirectional socket state.
- It is enough for "live-feeling" bingo if call cadence is seconds, not milliseconds.

Start rule:

- A scheduled game starts automatically at its scheduled time if more than 5 players are present.
- Winning pattern details become visible to players when the game starts, not during pre-game review/invite.

### 9. Automatic Claim Validation

Purpose:

- Remove host manual checking.

Behavior:

- Player clicks/announces Bingo in app.
- Backend validates assigned card against marked cells, called words, and active pattern rules.
- If valid, backend records winner placement automatically.
- Winner and host are notified.
- Game continues until top 3 winner placements are filled.

Claim validation should check:

- Player belongs to run.
- Player is using assigned card.
- Game is live.
- Pattern is active/allowed.
- Pattern exists on the player card.
- Every non-free-space cell in the pattern is marked.
- Every non-free-space word in the pattern has been called.

### 10. Prize Fulfillment

Purpose:

- Send winner reward emails automatically.
- Support voucher-claim links when the provider is known.

Recommended flow:

1. Winner confirmed.
2. Backend creates reward fulfillment job.
3. Worker calls gift card provider or reserves a preloaded code.
4. Worker emails winner.
5. Worker emails host/admin winner summary.
6. Game run stores fulfillment status.

Prize statuses:

- `not_required`
- `pending`
- `reserved`
- `sent`
- `failed`
- `manual_review`

Reward rule:

- The same player may receive multiple rewards in one game if they win multiple times.
- Reward amount/value can vary by pattern, placement, or game template settings.

Important:

- Gift cards are money-like. Fulfillment should have strong logs, retry rules, and admin visibility even if managers do not need a full audit UI.

### 11. Admin History

Purpose:

- Give admins useful oversight without overwhelming managers.

Admin should see:

- All game runs.
- Host.
- Audience count.
- Start/end time.
- Winners.
- Prize fulfillment status.
- Failed invites/emails.
- Failed AI generation/caller jobs.
- Host privilege requests.
- Immutable log entries for reward fulfillment, voice profile actions, and AI content generation/approval.

Managers should see:

- Their own recurring templates.
- Upcoming game runs.
- Current game.
- Winner summary.
- Delivery status at a simple level.

## Data Model Additions

The earlier Go backend plan still applies, but this autonomous version needs more tables.

Add:

- `host_privilege_requests`
- `game_templates`
- `game_template_schedules`
- `game_runs`
- `audience_sources`
- `audience_members`
- `run_invites`
- `invite_deliveries`
- `prompt_libraries`
- `prompts`
- `content_generation_jobs`
- `word_sets`
- `word_set_words`
- `voice_profiles`
- `voice_consents`
- `voice_profile_events`
- `caller_scripts`
- `caller_audio_assets`
- `caller_events`
- `reward_campaigns`
- `reward_fulfillments`
- `security_audit_events`
- `notification_jobs`
- `teams_installations`
- `teams_conversation_refs`
- `outbound_messages`

## Backend Services

Inside the Go backend, organize around these services:

- `auth`: Entra token verification and role mapping.
- `access`: host/admin privilege checks and approval workflow.
- `templates`: recurring game setup.
- `scheduler`: materialize future game runs.
- `audiences`: resolve mailing lists/groups into participant snapshots.
- `content`: prompt libraries, AI generation, word bank approval.
- `cards`: card generation and assignment.
- `game`: live game state and run lifecycle.
- `caller`: autonomous word calling, scripts, audio generation.
- `claims`: automatic validation and winner placement.
- `notifications`: Outlook and Teams delivery.
- `rewards`: gift card fulfillment.
- `summaries`: on-screen and email summaries.
- `admin`: cross-game admin visibility.

## Worker Jobs

Use durable queued jobs for anything that talks to external services or might need retry.

Jobs:

- `materialize_recurring_game`
- `resolve_audience`
- `generate_word_set`
- `review_deadline_check`
- `send_game_invites`
- `send_game_reminders`
- `open_lobby`
- `start_autonomous_game`
- `call_next_word`
- `generate_caller_audio`
- `post_teams_word_update`
- `validate_claim`
- `fulfill_reward`
- `send_winner_email`
- `send_host_summary`
- `complete_game_run`

Recommended Azure primitive:

- Azure Service Bus queue/topic for job messages.
- Azure Container Apps Jobs or Azure Functions timer triggers for scheduled work.

## Weekly Lifecycle

1. Admin approves host privileges.
2. Host creates recurring game template.
3. Host selects audience, recurrence, pattern rules, AI content mode, voice mode, and prize configuration.
4. Scheduler creates the next game run.
5. AI content job generates word set/cards from selected prompt.
6. Host optionally reviews and edits content.
7. Audience snapshot is resolved from mailing list/Teams/group.
8. Invites are sent through Outlook and/or Teams.
9. Players join from invite or sign-in dashboard.
10. System assigns each player a persisted card.
11. Game opens and autonomous caller begins.
12. Caller announces words and posts updates.
13. Players mark cards and claim Bingo.
14. Backend validates claims automatically.
15. Top 3 winners are recorded.
16. Winner emails and gift card workflow run.
17. Host/admin summary is sent.
18. Game history is available to admins.
19. Next weekly run is prepared automatically.

## API Areas To Add

### Host/Admin

- `GET /api/v1/admin/game-runs`
- `GET /api/v1/admin/host-requests`
- `POST /api/v1/admin/host-requests/{id}/approve`
- `POST /api/v1/admin/host-requests/{id}/reject`
- `POST /api/v1/host-requests`

### Templates

- `POST /api/v1/game-templates`
- `GET /api/v1/game-templates`
- `GET /api/v1/game-templates/{id}`
- `PATCH /api/v1/game-templates/{id}`
- `POST /api/v1/game-templates/{id}/pause`
- `POST /api/v1/game-templates/{id}/resume`
- `POST /api/v1/game-templates/{id}/generate-next-run`

### Content

- `POST /api/v1/prompt-libraries`
- `POST /api/v1/prompt-libraries/{id}/prompts`
- `POST /api/v1/game-runs/{id}/content/generate`
- `GET /api/v1/game-runs/{id}/content`
- `PATCH /api/v1/game-runs/{id}/content`
- `POST /api/v1/game-runs/{id}/content/approve`

### Invites

- `POST /api/v1/game-runs/{id}/audience/resolve`
- `POST /api/v1/game-runs/{id}/invites/send`
- `GET /api/v1/game-runs/{id}/invites`

### Live Game

- `GET /api/v1/game-runs/{id}/host-snapshot`
- `GET /api/v1/game-runs/{id}/player-snapshot`
- `GET /api/v1/game-runs/{id}/events`
- `POST /api/v1/game-runs/{id}/start`
- `POST /api/v1/game-runs/{id}/pause`
- `POST /api/v1/game-runs/{id}/resume`
- `POST /api/v1/game-runs/{id}/caller/next`
- `POST /api/v1/game-runs/{id}/caller/skip`
- `POST /api/v1/game-runs/{id}/claims`

### Rewards

- `GET /api/v1/game-runs/{id}/rewards`
- `POST /api/v1/game-runs/{id}/rewards/retry`
- `POST /api/v1/game-runs/{id}/summary/email`

## Revised Production V1 Order

The old prototype mindset was "make live bingo real." The Production V1 target is "make one automated weekly run production-ready."

### V1 Track 1: Internal Auth And Roles

- Entra auth skeleton.
- Admin/host/player role model.
- Host privilege request and approval workflow.
- CGI partner access baseline with per-game allowlist enforcement.

### V1 Track 2: Manual Template To Game Run

- Host creates a recurring game template.
- Backend materializes one game run manually.
- Host-managed allowed-player list.
- Invites can be mocked/logged first.
- Start threshold configured as more than 5 players.

### V1 Track 3: Real Game Engine

- Player join/rejoin.
- Persisted cards.
- Word calls.
- Automatic claim validation.
- Top 3 winners.
- On-screen summary.

### V1 Track 4: Automated Notifications

- Microsoft Graph email invite.
- Winner email.
- Host summary email.
- Basic Teams notification.

### V1 Track 5: AI Content Generation

- Prompt library.
- Generate word set from prompt.
- Host edit/approve.
- Review email and auto-approval fallback.
- Reuse word sets.

### V1 Track 6: Autonomous Caller

- Autonomous call cadence.
- AI caller text.
- Azure neural voice playback.
- Voice profile approval model.
- Admin-managed consented employee voice profiles.
- Host voice selection and auto mode.

### V1 Track 7: Rewards

- Reward provider adapter.
- Gift card fulfillment statuses.
- Winner gift card email.
- Retry/manual review path.

## Important Architecture Implications

1. The scheduler is core, not a nice-to-have.
   Recurring games are the main product. `game_templates` and `game_runs` should be built early.

2. Notifications need durable jobs.
   Email, Teams, AI generation, and gift cards will fail sometimes. The backend needs retries and status tracking.

3. AI generation needs approval states.
   Fully automatic can be a setting, but the data model must support draft, edited, approved, rejected, and reused content.

4. Voice cloning needs consent.
   Replicating employee voices is not just a technical feature. It needs explicit consent, profile status, admin approval, and a fallback voice.

5. Prize fulfillment needs stronger controls than normal email.
   Gift cards should have provider integration, fulfillment logs, and restricted admin access.

6. Realtime should serve the autonomous caller.
   The live layer is less about chat and more about syncing the conductor: current word, audio cue, card state, winner state.

7. Teams integration is two separate things.
   Teams notifications/messages are one capability. A Teams meeting with a live AI caller is another, harder capability. We should build in-app live audio first, then deepen Teams meeting integration.

## Remaining External Dependency

The main unresolved external dependency is the gift card/voucher provider.

Working assumption:

- Rewards are sent as voucher claim links by email.
- The backend should expose a provider adapter so the exact voucher vendor can be plugged in later.
- Reward fulfillment should still have statuses, retries, admin visibility, and immutable security/audit logs from the start.

## External References Checked

- Microsoft Graph `sendMail`: https://learn.microsoft.com/en-us/graph/api/user-sendmail
- Microsoft Graph online meetings: https://learn.microsoft.com/en-us/graph/api/application-post-onlinemeetings
- Microsoft Graph group members: https://learn.microsoft.com/en-us/graph/api/group-list-members
- Teams proactive messages: https://learn.microsoft.com/en-us/microsoftteams/platform/bots/how-to/conversations/send-proactive-messages
- Teams proactive app installation: https://learn.microsoft.com/en-us/microsoftteams/platform/graph-api/proactive-bots-and-messages/graph-proactive-bots-and-messages
- Azure AI Speech personal voice consent: https://learn.microsoft.com/en-us/azure/ai-services/speech-service/personal-voice-create-consent
- Azure AI Speech / Azure OpenAI speech: https://learn.microsoft.com/en-us/azure/ai-services/speech-service/openai-speech
- Azure Functions timer trigger: https://learn.microsoft.com/azure/azure-functions/functions-bindings-timer
- Azure Container Apps Jobs: https://learn.microsoft.com/en-us/azure/container-apps/jobs
- Azure Service Bus queues/topics: https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-queues-topics-subscriptions
- Azure Key Vault overview: https://learn.microsoft.com/en-us/azure/key-vault/general/overview
