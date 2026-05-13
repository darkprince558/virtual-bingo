# Technical Stack Proposal: Autonomous Real-Time Virtual Bingo Platform

Prepared for: Santiago Villasis (Manager)  
Prepared by: Anish Jami, Nathanial Ahluwalia  
Document status: Updated architecture draft  
Last updated: 2026-05-12

## 1. Executive Summary

This document proposes an updated enterprise-ready technology stack for an autonomous real-time Virtual Bingo platform.

The original demo frontend was built to show the product concept. The production system should be designed as a separate Azure-hosted platform that automates the weekly bingo lifecycle: recurring game creation, AI-assisted content generation, invite delivery, live in-app gameplay, AI voice calling, automatic winner validation, winner notifications, and future voucher/gift-card fulfillment.

The recommended architecture uses:

- Next.js frontend for players, hosts, and admins.
- Go backend as the deterministic source of truth for game state and winner validation.
- Python FastAPI AI service for content generation, caller scripts, and voice/audio workflows.
- Azure PostgreSQL for persistent data.
- Azure Service Bus for durable background jobs.
- Redis for realtime fanout/cache when scale requires it.
- Microsoft Entra ID for CGI/Microsoft identity.
- Microsoft Graph for Outlook/Teams delivery.
- Azure OpenAI and Azure AI Speech for AI content and voice.
- Azure Blob Storage for generated audio, exports, and artifacts.
- Azure Container Apps / Container Apps Jobs for API and worker deployment.
- Azure Key Vault, Managed Identities, Application Insights, and Log Analytics for secure operations.

The design keeps deterministic game decisions out of AI systems. AI can generate content, narration, scripts, and audio, but the Go backend owns game state, card assignment, claim validation, winner ordering, access control, and reward workflow state.

## 2. Business Context

The current manual bingo process includes:

- Generating cards through an external website.
- Emailing individual cards to participants.
- Hosting gameplay through Microsoft Teams.
- Manually calling words.
- Manually typing called words into chat.
- Manually validating winners.
- Manually tracking 1st, 2nd, and 3rd place.
- Manually sending prize notifications or gift cards.

The future platform should reduce manager workload as much as possible. Managers should configure a recurring game once, then let the system create weekly game runs, generate content, send invites, run the live in-app caller, validate winners, and notify winners/hosts automatically.

## 3. Product Direction

The production target is an autonomous weekly game operations platform.

Locked assumptions:

- The game happens inside the web app for now.
- Teams and Outlook are invite, reminder, notification, and re-entry channels.
- Access uses CGI/Microsoft identity through Microsoft Entra ID.
- The app can be broadly available to CGI partners.
- Each game has a host-managed allowlist of eligible players.
- Hosts can create games only if they have host privileges.
- Admins approve host privilege requests and can view all games.
- Scheduled games start only if more than 5 players have joined.
- Winning pattern is revealed when the game starts.
- AI-generated content requires host approval by default.
- If the host does not review in time, content is auto-approved so recurring games still run.
- Admins manage consented employee voice profiles.
- Hosts can select approved voices or use automatic voice selection.
- Players can win multiple times.
- The game tracks top 3 winners.
- Rewards will likely be voucher claim links by email, but provider details are still pending.
- Admin/security logs are required for reward, voice, and AI content actions.

## 4. Proposed Solution Overview

The platform consists of these major components:

1. Web frontend
   - Player game UI.
   - Host dashboard.
   - Admin dashboard.
   - Recurring game setup.
   - AI content review/editing.

2. Go backend
   - Auth and authorization.
   - Game templates and game runs.
   - Allowlists and player eligibility.
   - Card generation and assignment.
   - Live game state.
   - Word calling sequence.
   - Automatic claim validation.
   - Winner ordering.
   - Reward workflow state.
   - Admin/security audit logs.

3. Python AI service
   - AI word/card content generation.
   - Prompt library generation support.
   - AI caller scripts.
   - Azure AI Speech audio generation orchestration.
   - Host-assistant style commentary.

4. Worker/scheduler layer
   - Materialize weekly game runs.
   - Resolve allowed players.
   - Generate weekly content.
   - Send pre-game review emails.
   - Auto-approve content when needed.
   - Send invites/reminders.
   - Start games when threshold is met.
   - Drive autonomous caller jobs.
   - Send winner and host summary emails.
   - Trigger reward fulfillment adapters.

5. Azure platform services
   - PostgreSQL, Service Bus, Blob Storage, Key Vault, Container Apps, Application Insights, Log Analytics, and Azure Container Registry.

## 5. Recommended Technology Stack

| Category | Recommended Technology | Purpose |
|---|---|---|
| Frontend | React / Next.js | Player, host, and admin web interface |
| Frontend hosting | Azure Static Web Apps or Azure Container Apps | Host the web application |
| Backend language | Go | Deterministic game engine and API |
| Backend hosting | Azure Container Apps | Host containerized Go API |
| Realtime protocol | Server-Sent Events first, WebSockets later if needed | Live game updates |
| AI microservice | Python / FastAPI | AI content, narration scripts, voice workflows |
| AI content | Azure OpenAI | Prompt-based word/card generation and caller scripts |
| Voice / TTS | Azure AI Speech | AI voice calling and narration |
| Voice governance | Admin-managed voice profiles + consent records | Control replicated employee voices |
| Database | Azure Database for PostgreSQL Flexible Server | Persistent storage for users, games, cards, calls, claims, winners, schedules, rewards, and logs |
| Durable jobs | Azure Service Bus | Reliable async jobs, retries, and workflow messaging |
| Scheduled jobs | Azure Container Apps Jobs or Azure Functions timer triggers | Weekly materialization, reminders, and automation |
| Cache / realtime broker | Azure Cache for Redis | Realtime fanout, pub/sub, temporary state, scale-out coordination |
| Object storage | Azure Blob Storage | Generated audio, exports, archived files, optional generated card PDFs |
| Authentication | Microsoft Entra ID | Corporate single sign-on |
| Microsoft integration | Microsoft Graph | Outlook email, Teams notifications, future group/list sync |
| Authorization | Role-based access control | Admin, host, player, viewer permissions |
| Email | Microsoft Graph `sendMail` first; Azure Communication Services/SendGrid if needed | Invites, review emails, winner emails, host summaries |
| Prize fulfillment | Provider adapter | Future voucher claim links or gift-card vendor integration |
| Source control | GitHub | Repository management and code review |
| CI/CD | GitHub Actions | Automated build, test, and deployment pipelines |
| Containerization | Docker / Docker Buildx | Package frontend/backend/AI services |
| Container registry | Azure Container Registry | Store built container images |
| Local development | Docker Compose | Local frontend, Go backend, AI service, Postgres, Redis |
| Secrets management | Azure Key Vault | Secure API keys, credentials, and connection strings |
| Service identity | Managed Identities | Azure service-to-service authentication |
| Monitoring | Azure Monitor | Platform monitoring and alerting |
| Application monitoring | Azure Application Insights | Application telemetry and performance tracking |
| Logging | Azure Log Analytics | Centralized logs and querying |
| Infrastructure as Code | Azure Bicep or Terraform | Repeatable cloud infrastructure provisioning |
| API documentation | OpenAPI / Swagger | API contracts and integration visibility |
| Testing | Playwright, Go tests, Pytest, k6 | Frontend, backend, AI service, and load testing |

## 6. High-Level Architecture

```text
User Browser
  |
  | HTTPS + SSE/WebSocket
  v
Next.js Frontend
  |
  | REST APIs + live event stream
  v
Go Backend API  <------>  PostgreSQL
  |
  | enqueue durable jobs
  v
Azure Service Bus
  |
  v
Workers / Scheduler
  |        |           |
  |        |           +--> Microsoft Graph (Outlook / Teams)
  |        +--------------> Python FastAPI AI Service
  |                         |
  |                         +--> Azure OpenAI
  |                         +--> Azure AI Speech
  |
  +--> Blob Storage

Realtime scale path:
Go Backend API <------> Redis <------> other Go API instances

Supporting services:
- Microsoft Entra ID for authentication
- Azure Key Vault for secrets
- Azure Monitor / Application Insights / Log Analytics for observability
- Azure Container Registry for images
```

## 7. Component Responsibilities

### 7.1 Frontend Application

The frontend provides the user experience for players, hosts, and admins.

Responsibilities:

- Microsoft sign-in entry points.
- Player lobby and game card UI.
- Live called-word display.
- AI caller text/audio display.
- Bingo mark and claim interactions.
- Rejoin live game from invite/sign-in.
- Host recurring template setup.
- Host allowlist management.
- Host AI content review/editing.
- Host game summary.
- Admin host privilege approvals.
- Admin voice profile management.
- Admin game history and operational status.

### 7.2 Go Backend

The Go backend is the authoritative source of truth.

Responsibilities:

- Validate Entra identity tokens.
- Enforce admin/host/player permissions.
- Manage host privilege requests and approvals.
- Manage recurring game templates.
- Materialize game runs.
- Enforce per-game allowlists.
- Generate unique persisted bingo cards.
- Assign cards to eligible players.
- Manage called words and game progression.
- Broadcast live game updates.
- Track player marks.
- Validate Bingo claims automatically.
- Determine winner order.
- Track top 3 winners.
- Allow multiple wins by the same player.
- Store admin/security logs for reward, voice, and AI content actions.
- Expose REST APIs and live event streams.

The backend must remain deterministic and auditable. Winner validation, game state, card assignment, and reward workflow state should not be delegated to AI.

### 7.3 Python AI Agent Microservice

The AI service provides non-authoritative enhancement features.

Responsibilities:

- Generate weekly word sets from prompts.
- Generate caller scripts and descriptions.
- Generate or coordinate Azure AI Speech audio.
- Support prompt libraries and random prompt selection.
- Support host-assistant commentary.
- Return structured outputs for Go backend approval/storage.

The AI service should not decide winners, mutate card ownership, or directly send rewards.

### 7.4 Worker And Scheduler Layer

The worker layer performs autonomous operations that may fail or need retry.

Jobs:

- `materialize_recurring_game`
- `resolve_allowed_players`
- `generate_word_set`
- `send_content_review_email`
- `auto_approve_content`
- `send_game_invites`
- `send_game_reminders`
- `open_lobby`
- `start_game_if_threshold_met`
- `call_next_word`
- `generate_caller_audio`
- `post_teams_update`
- `send_winner_email`
- `send_host_summary`
- `fulfill_reward`
- `complete_game_run`

Use Azure Service Bus for durable job messages and retries.

### 7.5 PostgreSQL Database

PostgreSQL stores all durable application data.

Core entities:

- Users
- Roles
- Host privilege requests
- Game templates
- Template schedules
- Game runs
- Allowed players
- Invite deliveries
- Prompt libraries
- Prompts
- Content generation jobs
- Word sets
- Word set words
- Bingo cards
- Bingo card cells
- Called words
- Player marks
- Bingo claims
- Winners
- Voice profiles
- Voice consents
- Caller scripts
- Caller audio assets
- Reward fulfillments
- Notification jobs
- Admin/security audit events

### 7.6 Redis

Redis is useful for realtime scale and low-latency fanout.

Use cases:

- Pub/sub between Go API instances.
- Realtime event distribution.
- Temporary presence/connection state.
- Rate-limit counters.

Redis should not replace PostgreSQL as the source of truth, and it should not replace Service Bus for durable jobs.

### 7.7 Azure Blob Storage

Blob Storage stores generated or exported assets.

Potential assets:

- Generated AI caller audio.
- Voice consent recordings or references, depending on compliance rules.
- Game result exports.
- Archived summaries.
- Optional card PDFs/images.

### 7.8 Microsoft Graph

Microsoft Graph should be the main Microsoft 365 integration layer.

Use cases:

- Outlook game invite emails.
- Pre-game review emails.
- Winner emails.
- Host summary emails.
- Teams channel/proactive bot messages.
- Future group/list membership sync.

The actual game should remain in-app for now. Teams meeting participation by the AI caller is deferred.

## 8. Authentication And Authorization

Authentication should use Microsoft Entra ID.

Authorization should use role-based access control.

| Role | Capabilities |
|---|---|
| Admin | Manage users, approve hosts, view all games, manage voice profiles, view security/audit logs, configure system settings |
| Host | Create recurring templates, manage allowed players, review AI content, select voice mode, view own games and winners |
| Player | Join allowed games, view assigned card, mark squares, claim Bingo, rejoin live games |
| Viewer | Observe game progress and leaderboard only |

Access rule:

- A CGI/Microsoft user can sign in broadly, but joining a specific game requires matching that game run's allowlist.

## 9. Voice And AI Governance

Voice replication requires explicit consent.

Backend model should include:

- Voice profile owner.
- Consent record.
- Consent recording/statement metadata.
- Admin approval status.
- Active/inactive state.
- Removal/deactivation history.
- Usage history.

Hosts can:

- Select an approved voice.
- Choose auto mode.
- Use default Azure neural voices.

Hosts cannot:

- Upload arbitrary employee voices.
- Use unapproved/inactive voice profiles.
- Bypass consent.

## 10. Prize And Voucher Workflow

The exact voucher provider is still pending.

Architecture rule:

- Implement rewards through a provider adapter.
- Store reward fulfillment state from the beginning.
- Do not hard-code a specific gift card vendor.

Likely future flow:

1. Winner is confirmed automatically.
2. Backend creates reward fulfillment job.
3. Worker calls voucher provider or reserves claim link.
4. Winner email is sent.
5. Host summary email is sent.
6. Fulfillment status is stored.

Reward statuses:

- `not_required`
- `pending`
- `reserved`
- `sent`
- `failed`
- `manual_review`

The same player can receive multiple rewards in the same game. Reward value can vary by placement, pattern, or template settings.

## 11. DevOps And Deployment Approach

Recommended repository structure:

```text
virtual-bingo/
  frontend/
  backend-go/
  ai-service/
  infrastructure/
  docs/
  docker-compose.yml
  README.md
```

Recommended CI/CD flow:

```text
Developer Push / Pull Request
  |
  v
GitHub Actions
  |
  | lint + typecheck frontend
  | run frontend tests
  | run Go tests
  | run Python tests
  | build containers
  | run migration checks
  | optional E2E smoke tests
  v
Azure Container Registry
  |
  v
Azure Deployments
  |
  | Frontend -> Azure Static Web Apps or Container Apps
  | Backend -> Azure Container Apps
  | AI Service -> Azure Container Apps
  | Workers -> Container Apps Jobs / Azure Functions
```

## 12. Security Considerations

Security should be built in from the start.

Required controls:

- Microsoft Entra ID authentication.
- Role-based access control.
- Per-game allowlist enforcement.
- Azure Key Vault for secrets.
- Managed Identities for Azure service access.
- HTTPS-only communication.
- Environment-based configuration.
- No secrets committed to GitHub.
- Input validation on all APIs.
- Rate limiting on auth, invite, claim, and host/admin endpoints.
- Immutable security logs for rewards, voice, and AI content.
- Restricted admin and host permissions.
- Consent enforcement for voice profiles.

Sensitive values:

- Database credentials.
- Redis connection strings.
- Azure OpenAI credentials/config.
- Azure AI Speech credentials/config.
- Microsoft Graph app credentials, if required.
- Voucher provider credentials.
- Email service credentials.
- Entra ID configuration/secrets, if required.

## 13. Monitoring And Operations

Recommended monitoring stack:

- Azure Monitor.
- Azure Application Insights.
- Azure Log Analytics.
- Azure Monitor Alerts.

Recommended metrics and logs:

- Active game runs.
- Active players per game.
- SSE/WebSocket connection count.
- Disconnects and reconnects.
- Backend response latency.
- Failed API requests.
- Failed auth attempts.
- AI generation failures.
- TTS generation failures.
- Email delivery failures.
- Invite delivery failures.
- Claim validation events.
- Winner confirmation events.
- Reward fulfillment statuses.
- Voice profile changes.
- Container health and restarts.
- Worker job retries/dead-letter messages.

## 14. Testing Strategy

The platform needs automated testing across frontend, backend, AI service, jobs, and realtime flows.

| Test Type | Tooling | Purpose |
|---|---|---|
| Frontend tests | Playwright or Cypress | Validate user flows and UI behavior |
| Backend unit tests | Go testing package | Validate card generation, game state, claim validation, scheduling rules |
| Backend integration tests | Go + test database | Validate API/database behavior |
| AI service tests | Pytest | Validate AI service contracts and fallbacks |
| API tests | Bruno, Postman, or generated OpenAPI tests | Validate REST API contracts |
| Load tests | k6 | Validate live game and 100-player target |
| Worker tests | Go/Python test suites | Validate retry/idempotency behavior |

Key test cases:

- Unique card generation.
- Assigned card ownership enforcement.
- Allowed-player enforcement.
- Duplicate called-word prevention.
- Scheduled start only when more than 5 players joined.
- Pattern reveal at game start.
- Valid bingo pattern detection.
- Invalid claim rejection.
- Multiple wins by same player.
- Correct top 3 winner ordering.
- Player reconnection handling.
- Realtime event delivery.
- Email notification triggering.
- Review auto-approval.
- Voice profile consent enforcement.
- Reward fulfillment status transitions.
- Admin/security log creation.

## 15. Key Design Recommendations

1. Keep Go as the source of truth.
   Winner validation, game state, ranking, card assignment, and reward workflow state should be deterministic and auditable.

2. Treat AI as an enhancement layer.
   AI should generate content, caller scripts, descriptions, and voice assets. It should not decide winners or mutate authoritative game state.

3. Build recurring operations early.
   The core product is weekly autonomous operation, so `game_templates`, schedules, game runs, allowlists, and jobs should be built before advanced UI polish.

4. Use Service Bus for durable workflows.
   Redis is good for realtime/cache, but emails, AI jobs, rewards, and scheduled automation need durable queues and retries.

5. Use Azure-native integrations first.
   Azure OpenAI, Azure AI Speech, Key Vault, Managed Identities, Container Apps, Service Bus, and Application Insights fit the target environment.

6. Add consent and security logs from the beginning.
   Voice profiles and rewards are sensitive. The model should include consent, admin actions, immutable logs, and deactivation/removal history.

7. Keep Teams integration pragmatic.
   In-app gameplay comes first. Teams and Outlook should send links, updates, and notifications. A full Teams-meeting AI caller can be a later phase.

## 16. Final Summary

The updated stack is appropriate for an enterprise-grade autonomous internal Virtual Bingo platform.

It supports:

- Modern web-based gameplay.
- Autonomous recurring weekly game setup.
- Corporate identity.
- Host/admin access control.
- Per-game allowlists.
- Reliable realtime gameplay.
- Deterministic winner validation.
- AI-assisted card/word generation.
- Consented AI voice calling.
- Automated communications.
- Future voucher/gift-card fulfillment.
- Secure secrets management.
- CI/CD and deployment automation.
- Monitoring and operational visibility.

The recommended implementation approach is:

1. Build identity, roles, templates, allowlists, and game runs.
2. Build the deterministic Go game engine.
3. Add notification jobs and Microsoft Graph delivery.
4. Add AI content generation and review/auto-approval.
5. Add autonomous caller and Azure AI Speech.
6. Add reward provider integration when the voucher provider is known.
