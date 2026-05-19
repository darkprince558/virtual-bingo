import { Game, CalledWord } from '../types/game';
import { Player, LeaderboardEntry, BingoCellData } from '../types/player';

export const mockPlayers: Player[] = [
  { id: 'p1', name: 'Alice Smith', state: 'Playing', connectionState: 'Connected' },
  { id: 'p2', name: 'Bob Johnson', state: 'Playing', connectionState: 'Connected' },
  { id: 'p3', name: 'Charlie Davis', state: 'Playing', connectionState: 'Connected' },
  { id: 'p4', name: 'Diana Prince', state: 'Playing', connectionState: 'Connected' },
  { id: 'p5', name: 'Evan Wright', state: 'Playing', connectionState: 'Connected' },
];

export const mockCalledWords: CalledWord[] = [
  { id: 'w1', word: 'Synergy', calledAt: new Date(Date.now() - 120000).toISOString() },
  { id: 'w2', word: 'Bandwidth', calledAt: new Date(Date.now() - 90000).toISOString() },
  { id: 'w3', word: 'Deep Dive', calledAt: new Date(Date.now() - 60000).toISOString() },
  { id: 'w4', word: 'Circle Back', calledAt: new Date(Date.now() - 30000).toISOString() },
];

export const mockCurrentWord: CalledWord = {
  id: 'w5',
  word: 'Action Item',
  calledAt: new Date().toISOString(),
};

export const mockGame: Game = {
  id: 'g1',
  code: 'BINGO-2024',
  status: 'Live',
  hostName: 'Admin Team',
  connectedPlayers: 42,
  calledWords: mockCalledWords,
  currentWord: mockCurrentWord,
  claims: [
    {
      id: 'c1',
      playerId: 'p2',
      playerName: 'Bob Johnson',
      pattern: 'Line',
      status: 'Pending',
      claimedAt: new Date().toISOString(),
    }
  ],
  activityEvents: [
    {
      id: 'a1',
      label: 'Game started',
      detail: 'Admin Team moved the session to Live.',
      createdAt: new Date(Date.now() - 180000).toISOString(),
      tone: 'success',
    },
    {
      id: 'a2',
      label: 'Word called',
      detail: 'Action Item is now visible to all players.',
      createdAt: new Date(Date.now() - 45000).toISOString(),
      tone: 'neutral',
    },
    {
      id: 'a3',
      label: 'Claim received',
      detail: 'Bob Johnson submitted a line claim for review.',
      createdAt: new Date().toISOString(),
      tone: 'warning',
    },
  ]
};

export const mockBingoCard: BingoCellData[] = [
  { id: 'c1', word: 'Synergy', isMarked: true },
  { id: 'c2', word: 'Touch Base', isMarked: false },
  { id: 'c3', word: 'Bandwidth', isMarked: true },
  { id: 'c4', word: 'Pivot', isMarked: false },
  { id: 'c5', word: 'Deep Dive', isMarked: true },
  { id: 'c6', word: 'Action Item', isMarked: true },
  { id: 'c7', word: 'Aligned', isMarked: false },
  { id: 'c8', word: 'Agile', isMarked: false },
  { id: 'c9', word: 'Buy-in', isMarked: false },
  { id: 'c10', word: 'Deliverable', isMarked: false },
  { id: 'c11', word: 'Ecosystem', isMarked: false },
  { id: 'c12', word: 'Friction', isMarked: false },
  { id: 'c13', word: 'FREE SPACE', isMarked: true },
  { id: 'c14', word: 'Granular', isMarked: false },
  { id: 'c15', word: 'Hard Stop', isMarked: false },
  { id: 'c16', word: 'Ideate', isMarked: false },
  { id: 'c17', word: 'Leverage', isMarked: false },
  { id: 'c18', word: 'Moving Forward', isMarked: false },
  { id: 'c19', word: 'On My Radar', isMarked: false },
  { id: 'c20', word: 'Circle Back', isMarked: true },
  { id: 'c21', word: 'Pain Point', isMarked: false },
  { id: 'c22', word: 'Ping Me', isMarked: false },
  { id: 'c23', word: 'Reach Out', isMarked: false },
  { id: 'c24', word: 'Scalable', isMarked: false },
  { id: 'c25', word: 'Value Add', isMarked: false },
];

export const mockLeaderboard: LeaderboardEntry[] = [
  { placement: 1, player: mockPlayers[0], wordsMatched: 12 },
  { placement: 2, player: mockPlayers[1], wordsMatched: 10 },
  { placement: 3, player: mockPlayers[2], wordsMatched: 9 },
  { placement: 4, player: mockPlayers[3], wordsMatched: 7 },
  { placement: 5, player: mockPlayers[4], wordsMatched: 5 },
];
