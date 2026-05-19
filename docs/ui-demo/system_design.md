# Virtual Bingo Game UI Design Specification

## 1. Product Overview

The Virtual Bingo Game is an internal corporate web application designed to replace the current manual virtual bingo process.

The application will centralize the bingo experience into one real-time web interface for hosts and players. It should support automated card display, real-time word calling, player card marking, Bingo claim submission, winner validation, leaderboard tracking, and game summary visibility.

The UI should be clean, professional, minimal, and easy to use during a live workplace event.

---

## 2. Design Goal

The interface should feel like:

> A polished internal corporate tool with light, fun game-night energy.

The application should be professional enough for manager review and workplace use, while still feeling engaging and enjoyable for players.

---

## 3. Target Users

### Host

The host creates and manages the game session.

Host needs:

- Start, pause, and end the game
- Call the next word
- See connected players
- Review Bingo claims
- View winners and leaderboard
- Monitor game activity

### Player

The player joins a game and plays Bingo.

Player needs:

- Join using a game code
- View assigned Bingo card
- See the current called word
- Mark card cells
- Claim Bingo
- View leaderboard and winners

### Admin / Viewer

Admins or viewers may need visibility into game results, audit history, or session summaries.

---

## 4. Visual Style

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
- Dark, overloaded layouts
- Too many colors
- Random decorative elements that do not support gameplay

---

## 5. Color Direction

Use a light, corporate-friendly palette.

### Recommended Colors

- Background: very light gray or off-white
- Surface cards: white
- Primary accent: blue or indigo
- Success state: green
- Warning state: amber
- Error state: red
- Neutral text: dark gray / near-black

### Usage Guidelines

- Use the primary accent for main actions, current word emphasis, and active game states.
- Use success colors for confirmed Bingo claims and winners.
- Use warning colors for pending claims or paused games.
- Use error colors for rejected Bingo claims or connection issues.
- Do not rely only on color to communicate state. Use text, icons, and labels as well.

---

## 6. Typography

Use a clean sans-serif font.

Recommended hierarchy:

- Page title: large and bold
- Section headers: medium and semibold
- Current called word: very large and bold
- Bingo card cells: readable and medium weight
- Supporting text: smaller and neutral
- Status labels: small but clear

The current called word and Bingo card should be readable from a distance or during screen sharing.

---

## 7. Layout Principles

### General Layout

Use a card-based layout with clear spacing.

Preferred structure:

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

The Bingo card should be the visual center of the player experience.

### Host Dashboard Priority

The host dashboard should prioritize:

1. Game status
2. Current called word
3. Game controls
4. Bingo claim queue
5. Player list
6. Leaderboard
7. Activity feed

The host should be able to run the game without searching through the interface.

---

## 8. Required Screens

## 8.1 Landing / Login Screen

Purpose:

Allow users to access or join the game.

Required elements:

- App name: Virtual Bingo
- Short product description
- Sign-in button placeholder
- Join game option
- Clean centered layout

Notes:

Authentication will later use Microsoft Entra ID, but the UI prototype can use mock login behavior.

---

## 8.2 Game Lobby Screen

Purpose:

Allow players to join a game before it begins.

Required elements:

- Game code
- Player name
- Waiting room status
- Connected player count
- Host information
- Game status: Waiting / Starting Soon / Live
- Join button

---

## 8.3 Player Bingo Game Screen

Purpose:

Main gameplay screen for players.

Required elements:

- Top bar with app name, game code, player name, and game status
- Current called word display
- AI caller description or message
- 5x5 Bingo card
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

## 8.4 Host Dashboard Screen

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

## 8.5 Winner Confirmation Modal

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

- Excessive confetti
- Loud colors
- Childish graphics
- Distracting animations

---

## 8.6 Game Summary Screen

Purpose:

Show final game results.

Required elements:

- Final winners
- Total players
- Total words called
- Game duration
- Prize notification status placeholder
- Export/share results placeholder
- Return to dashboard button

---

## 9. Component Architecture

Recommended frontend components:

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
