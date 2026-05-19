export type GameState = 'Waiting' | 'Starting Soon' | 'Lobby Open' | 'Live' | 'Paused' | 'Finished' | 'Cancelled' | 'Failed';

import { Player } from './player';

export type BingoPattern = 'Line' | 'Four Corners' | 'Full House';

export interface BingoClaim {
  id: string;
  playerId: string;
  playerName: string;
  pattern: BingoPattern;
  status: 'Pending' | 'Valid' | 'Invalid' | 'Confirmed';
  claimedAt: string;
}

export interface CalledWord {
  id: string;
  word: string;
  calledAt: string;
}

export interface ActivityEvent {
  id: string;
  label: string;
  detail: string;
  createdAt: string;
  tone?: 'neutral' | 'success' | 'warning' | 'danger';
}

export interface GameSummary {
  totalPlayers: number;
  totalWordsCalled: number;
  gameDurationMinutes: number;
  winners: {
    placement: 1 | 2 | 3;
    player: Player;
    pattern: BingoPattern;
  }[];
}

export interface Game {
  id: string;
  code: string;
  status: GameState;
  hostName: string;
  connectedPlayers: number;
  calledWords: CalledWord[];
  currentWord: CalledWord | null;
  claims: BingoClaim[];
  activityEvents: ActivityEvent[];
}
