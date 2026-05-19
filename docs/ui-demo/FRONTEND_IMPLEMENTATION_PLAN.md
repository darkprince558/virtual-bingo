# Virtual Bingo - Frontend Implementation Plan

## 1. Executive Summary
This document serves as the master blueprint for rebuilding the Virtual Bingo UI. The goal is to replace the current demo UI with a highly engaging, compelling, and fun interface that employees will enjoy using. The application will serve as the frontend for a Go-based autonomous Bingo backend.

## 2. Design Language: The "Headspace" Aesthetic
We will use a design language inspired by Headspace to make the corporate game feel un-intimidating, relaxing, and fun.

*   **Soft Geometry:** Aggressive use of rounded corners (`rounded-2xl`, `rounded-full`). No sharp edges.
*   **Color Palette:** Warm, playful color blocks. Signature warm oranges, soft mint greens (for success), warm yellows, and calming lavenders on a soft off-white background (`#FAF9F6`).
*   **Typography:** **Nunito** or **Outfit** (via `next/font/google`). These open-source fonts have soft, rounded letterforms mimicking Headspace's Aperçu.
*   **Micro-Interactions:** Using `framer-motion` to create "soft bounces" when interacting with buttons and bingo cells (e.g., squish and bounce back on tap).
*   **Icons:** Soft, clean icons using `lucide-react`.

### Inspiration & References
*   **GitHub/Reddit Clones:** Draw architectural patterns from open-source React/Tailwind Headspace clones (e.g., `nicolastorgato/headspace-ui-clone`, `shadcn/ui` structures).
*   **Google Stitch Principles:** Utilize AI-native generation principles (like maintaining strict design tokens in a central place) to ensure consistency across the generated components.

## 3. Layout Strategy: Mobile-Compatible
*   **Desktop Primary:** The application focuses on large screens first. The Host Dashboard and Player Screen will use rich, side-by-side panels.
*   **Mobile Fallback (Compatible):** When accessed on mobile, the UI will stack gracefully. Panels will collapse into accessible menus or a pill-shaped bottom navigation bar to keep the Bingo Card prominent.

## 4. Tech Stack (Already configured in `apps/frontend-demo`)
*   **Framework:** Next.js 15 (App Router)
*   **UI Library:** React 19
*   **Styling:** Tailwind CSS v4
*   **Animations:** Framer Motion (`motion`)
*   **Icons:** Lucide React

---

## 5. Major Screens to Build

### A. Landing / Login Screen (`app/page.tsx`)
*   **Purpose:** Entry point for players and hosts.
*   **Key Elements:**
    *   Large, colorful, playful graphic/logo.
    *   Input field for Game Code and Player Name.
    *   Prominent, pill-shaped "Join Game" button.
    *   Mock login functionality for the prototype phase.

### B. Game Lobby Screen (`app/lobby/[gameCode]/page.tsx`)
*   **Purpose:** Waiting room before the host starts the game.
*   **Key Elements:**
    *   Game Code display.
    *   Grid/List of connected players represented as colorful, rounded bubbles or avatars.
    *   Subtle idle animations (e.g., avatars gently bobbing) to indicate live presence.
    *   Status indicator: "Waiting for Host" or "Starting Soon".

### C. Player Bingo Game Screen (`app/game/[gameCode]/page.tsx`)
*   **Purpose:** The core gameplay interface for employees.
*   **Key Elements:**
    *   **Current Call Display:** A massive, colorful bubble at the top displaying the current called word.
    *   **The Bingo Card:** A 5x5 grid of chunky, rounded squares. Marking a cell triggers a color fill and a satisfying bouncy animation.
    *   **Side Panel (Desktop) / Bottom Nav (Mobile):** Contains the auto-updating Top 3 Leaderboard and a scrolling feed of Recently Called Words.
    *   **Claim Bingo Action:** A vibrant, floating action button (FAB) that pulses when the player is close to a Bingo.

### D. Host Dashboard Screen (`app/host/[gameCode]/page.tsx`)
*   **Purpose:** Control center for the game organizer.
*   **Key Elements:**
    *   **Control Panel:** High-contrast buttons to "Start Game", "Call Next Word", "Pause", and "End Game".
    *   **Player List:** Real-time view of connected players.
    *   **Claim Queue:** A dedicated section to instantly review, approve, or reject Bingo claims.
    *   **Activity Feed:** A log of all game events.

### E. Winner Confirmation Modal (Component)
*   **Purpose:** To celebrate a valid Bingo claim.
*   **Key Elements:**
    *   Full-screen or large centered overlay.
    *   Warm colors and a friendly "Winner!" graphic (avoiding harsh confetti).
    *   Displays 1st, 2nd, or 3rd place.

### F. Game Summary Screen (`app/summary/[gameCode]/page.tsx`)
*   **Purpose:** Post-game wrap-up.
*   **Key Elements:**
    *   Podium/List of final winners.
    *   Total duration and words called.
    *   Return to dashboard button.

---

## 6. Implementation Phasing for the Next Agent

When starting the new chat, the AI agent should follow these phases:

### Phase 1: Foundation
1. Update `app/layout.tsx` to load the **Nunito** font.
2. Define the Headspace color palette and shadow variables in `app/globals.css`.
3. Create core reusable UI components (`SoftButton`, `SoftCard`, `BingoCell`) using Tailwind and Framer Motion.
4. Scaffold mock data models (`lib/mockGameData.ts`).

### Phase 2: Pre-Game UI
1. Build the Landing Screen (`app/page.tsx`).
2. Build the Lobby Screen (`app/lobby/[gameCode]/page.tsx`).

### Phase 3: Core Game UI
1. Build the Host Dashboard (`app/host/[gameCode]/page.tsx`) focusing on the desktop layout.
2. Build the Player Screen (`app/game/[gameCode]/page.tsx`), ensuring the Bingo Card and animations feel perfect, and that it stacks correctly on mobile.

### Phase 4: Polish
1. Implement the Winner Confirmation Modal.
2. Build the Game Summary Screen.
3. Perform a final pass to ensure all tap targets are mobile-friendly and animations are smooth.
