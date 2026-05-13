# Virtual Bingo Game UI Design Specification

## 1. Product Overview

The Virtual Bingo Game is an internal corporate web application designed to replace a manual virtual bingo process.

The application should support:

- Automated bingo card display
- Real-time word calling
- Player card marking
- Bingo claim submission
- Automated winner validation
- Live leaderboard tracking
- AI-assisted caller messages
- Host game controls
- Game summary and audit visibility

The UI should be clean, professional, minimal, and easy to use during a live team event.

---

## 2. Design Goal

The interface should feel like:

> A polished internal corporate tool with light, fun game-night energy.

The UI should be suitable for manager demos and workplace use while still making the game feel engaging.

---

## 3. Visual Style

### Desired Feel

- Clean
- Modern
- Minimal
- Professional
- Friendly
- Slightly playful
- Easy to understand quickly

### Avoid

- Casino-style visuals
- Childish cartoon styling
- Heavy gradients
- Excessive animations
- Cluttered dashboards
- Dark overloaded layouts
- Too many colors
- Random decorative elements that do not support gameplay

---

## 4. Suggested Color Direction

Use a light, corporate-friendly palette.

### Primary Colors

- Background: very light gray or off-white
- Surface cards: white
- Primary accent: blue or indigo
- Success state: green
- Warning state: amber
- Error state: red
- Neutral text: dark gray / near-black

### Usage

- Primary accent should be used for main buttons, current word emphasis, and active game states.
- Success color should be used for confirmed Bingo claims and winners.
- Warning color should be used for pending claims or paused games.
- Error color should be used for rejected Bingo claims or connection issues.
- Do not rely only on color to communicate state. Use labels, icons, and text as well.

---

## 5. Typography

Use a clean sans-serif font.

Recommended hierarchy:

- Page title: large and bold
- Section headers: medium and semibold
- Current called word: very large and bold
- Bingo card cells: readable and medium weight
- Supporting text: smaller and neutral
- Status labels: small but clear

The current called word and bingo card should be readable from a distance or while screen-sharing.

---

## 6. Layout Principles

### General Layout

Use a card-based layout with clear spacing.

Preferred layout structure:

- Top navigation bar
- Main content area
- Supporting side panels
- Clear action buttons

### Player Screen Priority

The player screen should prioritize:

1. Current called word
2. Bingo card
3. Claim Bingo button
4. Recently called words
5. Leaderboard
6. AI caller messages

The bingo card should be the visual center of the player experience.

### Host Dashboard Priority

The host dashboard should prioritize:

1. Game status
2. Current called word
3. Game controls
4. Bingo claim queue
5. Player list
6. Leaderboard
7. Activity/audit feed

The host should be able to run the game without searching through the interface.

---

## 7. Required Screens

## 7.1 Landing / Login Screen

Purpose:  
Allow users to access the game.

Required elements:

- App name: Virtual Bingo
- Short description
- Sign-in button placeholder
- Join game option
- Clean centered layout

Notes:  
Authentication will later use Microsoft Entra ID, but the UI prototype can use mock login behavior.

---

## 7.2 Game Lobby Screen

Purpose:  
Allow players to join a game before it begins.

Required elements:

- Game code
- Player name
- Waiting room status
- List/count of connected players
- Host information
- Start status: Waiting / Starting Soon / Live
- Simple join button

---

## 7.3 Player Bingo Game Screen

Purpose:  
Main gameplay screen for players.

Required elements:

- Top bar with app name, game code, player name, and game status
- Current called word display
- AI caller description/message
- 5x5 bingo card
- Marked and unmarked cell states
- Claim Bingo button
- Recently called words feed
- Leaderboard
- AI caller/chat panel
- Connection status indicator

Important behavior:

- Marked cells should be visually clear.
- The current called word should be obvious immediately.
- The Claim Bingo button should be visible but not distracting.
- The leaderboard should show 1st, 2nd, and 3rd place clearly.

---

## 7.4 Host Dashboard Screen

Purpose:  
Allow the host to run and monitor the game.

Required elements:

- Game status badge
- Connected player count
- Current called word
- Start Game button
- Call Next Word button
- Pause Game button
- End Game button
- Bingo claim queue
- Player list
- Called word history
- Leaderboard
- Activity/audit feed

Important behavior:

- Host controls should be grouped together.
- Dangerous actions like End Game should look visually distinct.
- Bingo claims should be easy to review.
- Winner order should be clear.

---

## 7.5 Winner Confirmation Modal

Purpose:  
Display confirmed winner results.

Required elements:

- Winner name
- Placement: 1st, 2nd, or 3rd
- Winning pattern
- Confirmation message
- Continue Game button
- Subtle celebration effect

Style:  
Celebratory but professional.

Avoid:  
Over-the-top confetti, loud colors, or childish graphics.

---

## 7.6 Game Summary Screen

Purpose:  
Show the final game results.

Required elements:

- Final winners
- Total players
- Total words called
- Game duration
- Prize notification status placeholder
- Export/share results placeholder
- Return to dashboard button

---

## 8. Component Architecture

Recommended components:

```text
components/
  AppShell.tsx
  TopNav.tsx
  GameStatusBadge.tsx
  BingoCard.tsx
  BingoCell.tsx
  CurrentCallDisplay.tsx
  CalledWordsFeed.tsx
  Leaderboard.tsx
  AIHostPanel.tsx
  HostControls.tsx
  PlayerLobby.tsx
  PlayerList.tsx
  BingoClaimQueue.tsx
  WinnerModal.tsx
  GameSummary.tsx
```

Recommended support files:

```text
types/
  game.ts
  player.ts
  websocket.ts

lib/
  mockGameData.ts
  gameHelpers.ts
  websocketClient.ts
```

---

## 9. UI States

The UI should account for these states:

### Game States

- Waiting
- Live
- Paused
- Finished

### Player States

- Joined
- Waiting
- Playing
- Claimed Bingo
- Confirmed Winner
- Rejected Claim
- Disconnected

### Bingo Claim States

- Pending
- Valid
- Invalid
- Confirmed

### Connection States

- Connected
- Reconnecting
- Disconnected

---

## 10. Accessibility Requirements

The UI should follow basic accessibility expectations:

- Use readable text sizes
- Maintain strong color contrast
- Do not rely only on color to show state
- Use visible focus states
- Make buttons keyboard accessible
- Use clear button labels
- Use semantic HTML where possible
- Add aria-labels for icon-only buttons
- Ensure marked bingo cells are visually and textually identifiable

Examples:

- A marked cell should show both a visual highlight and a check mark.
- Game status should use both color and text.
- Winner placement should be written clearly as “1st Place”, “2nd Place”, or “3rd Place”.

---

## 11. Animation Guidelines

Use animations sparingly.

Good animation examples:

- Subtle pulse on the current called word
- Smooth cell marking transition
- Light winner modal entrance
- Small celebration effect for confirmed winner
- Gentle loading states

Avoid:

- Constant motion
- Distracting background animations
- Excessive confetti
- Spinning elements
- Audio autoplay without controls

---

## 12. Responsive Design

The primary target should be desktop and tablet.

### Desktop

Use a multi-column layout:

- Main game area in center
- Supporting panels on right
- Host controls grouped clearly

### Tablet

Stack panels where needed, but keep the bingo card and current word highly visible.

### Mobile

Mobile support is nice-to-have for MVP. If supported, prioritize:

- Current word
- Bingo card
- Claim button
- Recently called words

---

## 13. MVP Scope

The first UI version should include:

- Landing/login mock screen
- Game lobby
- Player bingo screen
- Host dashboard
- Winner modal
- Game summary
- Mock game data
- Mock leaderboard
- Mock called words
- Mock Bingo claims

Do not include in the first UI prototype:

- Real authentication
- Real backend integration
- Real WebSocket connection
- Real email delivery
- Real AI API calls
- Full admin analytics
- Complex settings pages

---

## 14. Handoff Notes for Engineering

The UI should be built so that mock data can later be replaced with real backend data.

Important integration points:

- Current called word will come from WebSocket events
- Player card will come from backend-assigned card data
- Marked cells may be stored locally and/or sent to backend
- Bingo claim will call backend validation
- Leaderboard will update from backend game state
- Host controls will call backend REST/WebSocket actions
- AI caller messages will come from the AI service through backend events

The frontend should not decide final winners. Winner validation should be handled by the Go backend.

---

## 15. Design Success Criteria

The UI is successful if:

- A player immediately understands how to play
- The current called word is obvious
- The bingo card is easy to read and mark
- The host can run the game without confusion
- Bingo claims and winners are clearly shown
- The leaderboard feels automatic and trustworthy
- The app feels professional enough for workplace use
- The experience still feels fun and event-like

