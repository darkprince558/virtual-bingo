export type PlayerState = 'Joined' | 'Waiting' | 'Playing' | 'Claimed Bingo' | 'Confirmed Winner' | 'Rejected Claim' | 'Disconnected';
export type ConnectionState = 'Connected' | 'Reconnecting' | 'Disconnected';

export interface Player {
  id: string;
  name: string;
  state: PlayerState;
  connectionState: ConnectionState;
  score?: number;
  avatarUrl?: string;
}

export interface BingoCellData {
  id: string;
  word: string;
  isMarked: boolean;
  isFreeSpace?: boolean;
}

export interface LeaderboardEntry {
  placement: 1 | 2 | 3 | number;
  player: Player;
  wordsMatched: number;
}
