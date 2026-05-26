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

export interface GameRunSummaryResponse extends GameRunResponse {
  playerCount: number;
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
  reconnectNotice?: ReconnectNoticeResponse;
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
  callerAsset?: CallerAssetResponse;
}

export interface CallerAssetResponse {
  id: string;
  gameRunId: string;
  callDeckItemId: string;
  word: string;
  sequence: number;
  line: string;
  audioUrl?: string;
  storageKey?: string;
  voiceName?: string;
  provider: string;
  status: string;
  errorReason?: string;
}

export interface ClaimResponse {
  id: string;
  gameRunId: string;
  playerId: string;
  pattern: string;
  status: string;
  validationResult?: {
    valid?: boolean;
    reason?: string;
    missingWords?: string[];
    matchedCells?: string[];
    [key: string]: unknown;
  };
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

export interface GameSettingsResponse {
  id?: string;
  gameRunId: string;
  markingMode: string;
  allowPlayerMarkingModeChoice: boolean;
  showClaimReadiness: boolean;
  voiceClaimMode?: string;
  voiceClaimAutoplay?: boolean;
  callerMode?: string;
  themeMode?: string;
  themeId?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface ReconnectNoticeResponse {
  lastSeenAt: string;
  missedCalledWords: CalledWordResponse[];
}

export interface HostSnapshotResponse {
  gameRun: GameRunResponse;
  settings?: GameSettingsResponse;
  status: string;
  currentWord?: CalledWordResponse;
  currentCallerAsset?: CallerAssetResponse;
  appliedTheme?: unknown;
  winningPattern: string;
  playerCount: number;
  players: PlayerResponse[];
  calledWords: CalledWordResponse[];
  claims: ClaimResponse[];
  winners: WinnerResponse[];
}

export interface PlayerSnapshotResponse {
  gameRun: GameRunResponse;
  markingMode?: string;
  allowPlayerMarkingModeChoice?: boolean;
  showClaimReadiness?: boolean;
  status: string;
  currentWord?: CalledWordResponse;
  currentCallerAsset?: CallerAssetResponse;
  appliedTheme?: unknown;
  winningPattern: string;
  player: PlayerResponse;
  card?: CardResponse;
  calledWords: CalledWordResponse[];
  claims: ClaimResponse[];
  winners: WinnerResponse[];
  reconnectNotice?: ReconnectNoticeResponse;
}

export interface ClaimSubmissionResponse {
  claim: ClaimResponse;
  winner?: WinnerResponse;
}

export interface ClaimReadinessResponse {
  ready: boolean;
  supportedPatterns: string[];
  readyPatterns: string[];
  bestPattern: string;
  matchedCells: CardCellResponse[];
  missingCells: CardCellResponse[];
  reason: string;
}

export interface GameContentResponse {
  id: string;
  gameRunId: string;
  generationJobId?: string;
  status: string;
  topic: string;
  summary: string;
  words: string[];
  generatedWords?: string[];
  callerStyle?: string;
  themePrompt?: string;
  reviewWindowOpensAt?: string;
  reviewWindowClosesAt?: string;
  lockedAt?: string;
  lockedWordSetId?: string;
  generationProvider: string;
  generationError?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ActivityEventResponse {
  id: string;
  gameRunId: string;
  type: string;
  entityType?: string;
  entityId?: string;
  actorUserId?: string;
  payload?: Record<string, unknown>;
  sequence?: number;
  createdAt: string;
}

export interface AllowedPlayerResponse {
  id: string;
  gameRunId: string;
  email: string;
  displayName: string;
  source: string;
  createdAt: string;
}

export interface WordSetWordResponse {
  id: string;
  wordSetId: string;
  word: string;
  sortOrder: number;
  isActive: boolean;
  createdAt: string;
}

export interface WordSetResponse {
  id: string;
  name: string;
  status: string;
  source: string;
  words?: WordSetWordResponse[];
  createdAt: string;
  updatedAt: string;
}

export interface DeliveryAttemptResponse {
  id: string;
  gameRunId: string;
  channel: string;
  purpose: string;
  recipientEmail: string;
  subject: string;
  templateKey: string;
  bodyPreview: string;
  linkUrl: string;
  gameCode: string;
  status: string;
  errorReason?: string;
  sentAt?: string;
  createdAt: string;
}

export interface ThemeResponse {
  id: string;
  gameRunId?: string;
  name: string;
  summary: string;
  tokens: Record<string, unknown>;
  status: string;
  provider: string;
  createdAt: string;
  updatedAt: string;
}

export interface GameSummaryResponse {
  gameRun: GameRunResponse;
  playerCount: number;
  calledWordCount: number;
  currentWord?: CalledWordResponse;
  claims: ClaimResponse[];
  winners: WinnerResponse[];
  players: PlayerResponse[];
  calledWords: CalledWordResponse[];
  status: string;
}

export interface ClaimAcknowledgementResponse extends ActivityEventResponse {}

export interface GameEventResponse {
  id: string;
  gameRunId: string;
  type: string;
  entityId?: string;
  payload: any;
  sequence: number;
  createdAt: string;
}

export interface ChatMessageResponse {
  id: string;
  gameRunId: string;
  playerId?: string;
  senderEmail: string;
  senderDisplayName: string;
  body: string;
  status: 'visible' | 'deleted' | 'hidden';
  createdAt: string;
}
