export interface GameRunResponse {
  id: string;
  templateId?: string;
  hostUserId: string;
  wordSetId?: string;
  code: string;
  name: string;
  status: string;
  scheduledStartAt?: string;
  startedAt?: string;
  endedAt?: string;
  currentCalledWordId?: string;
  winningPattern?: string;
  allowedPlayerCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface PlayerResponse {
  id: string;
  gameRunId: string;
  userId?: string;
  email: string;
  displayName: string;
  icon?: string;
  avatarColor?: string;
  avatarLabel?: string;
  connectionState: string;
  state: string;
  joinedAt: string;
  lastSeenAt: string;
}

export interface CardCellResponse {
  id: string;
  rowIndex: number;
  colIndex: number;
  word: string;
  isFreeSpace: boolean;
  markedAt?: string;
}

export interface CardResponse {
  id: string;
  gameRunId: string;
  playerId: string;
  seed: string;
  cells: CardCellResponse[];
  createdAt: string;
}

export interface CalledWordResponse {
  id: string;
  gameRunId: string;
  wordSetWordId?: string;
  word: string;
  calledByUserId?: string;
  sequence: number;
  calledAt: string;
  createdAt: string;
}

export interface ClaimResponse {
  id: string;
  gameRunId: string;
  playerId: string;
  pattern: string;
  status: string;
  validationResult: any;
  claimedAt: string;
  reviewedByUserId?: string;
  reviewedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface WinnerResponse {
  id: string;
  gameRunId: string;
  playerId: string;
  claimId?: string;
  placement: number;
  pattern: string;
  confirmedAt: string;
  createdAt: string;
}

export interface HostSnapshotResponse {
  gameRun: GameRunResponse;
  status: string;
  currentWord?: CalledWordResponse;
  winningPattern: string;
  playerCount: number;
  players: PlayerResponse[];
  calledWords: CalledWordResponse[];
  claims: ClaimResponse[];
  winners: WinnerResponse[];
}

export interface PlayerSnapshotResponse {
  gameRun: GameRunResponse;
  status: string;
  currentWord?: CalledWordResponse;
  winningPattern: string;
  player: PlayerResponse;
  card?: CardResponse;
  calledWords: CalledWordResponse[];
  claims: ClaimResponse[];
  winners: WinnerResponse[];
}

export interface GameEventResponse {
  id: string;
  gameRunId: string;
  type: string;
  entityId?: string;
  payload: any;
  sequence: number;
  createdAt: string;
}
