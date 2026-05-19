// ─── Host Privilege Requests ───
export type HostRequestStatus = 'Pending' | 'Approved' | 'Rejected';

export interface HostPrivilegeRequest {
  id: string;
  requesterName: string;
  requesterEmail: string;
  requestedAt: string;
  reason: string;
  status: HostRequestStatus;
  reviewedAt?: string;
  reviewedBy?: string;
}

// ─── Voice Profiles ───
export type VoiceProfileStatus = 'Active' | 'Inactive' | 'Pending Consent' | 'Revoked';

export interface VoiceProfile {
  id: string;
  employeeName: string;
  employeeEmail: string;
  status: VoiceProfileStatus;
  consentRecordedAt?: string;
  approvedBy?: string;
  approvedAt?: string;
  lastUsedAt?: string;
  usageCount: number;
}

// ─── Game Run (Admin view) ───
export type GameRunStatus =
  | 'Draft'
  | 'Content Generating'
  | 'Content Review'
  | 'Scheduled'
  | 'Invites Sent'
  | 'Lobby Open'
  | 'Live'
  | 'Paused'
  | 'Finished'
  | 'Reward Fulfillment'
  | 'Complete'
  | 'Cancelled'
  | 'Failed';

export interface GameRunSummary {
  id: string;
  templateName: string;
  hostName: string;
  scheduledAt: string;
  status: GameRunStatus;
  playerCount: number;
  winnersCount: number;
}

// ─── Game Templates ───
export type ContentMode = 'AI Generated' | 'Manual' | 'Reuse Word Bank';
export type VoiceMode = 'Default Neural' | 'Employee Voice' | 'Auto Select';
export type RecurrenceRule = 'Weekly' | 'Bi-Weekly' | 'Monthly' | 'One-Time';

export interface GameTemplate {
  id: string;
  name: string;
  hostName: string;
  recurrence: RecurrenceRule;
  dayOfWeek: string;
  time: string;
  timezone: string;
  audienceSize: number;
  contentMode: ContentMode;
  voiceMode: VoiceMode;
  minPlayers: number;
  prizeEnabled: boolean;
  prizeValue?: string;
  nextRunAt?: string;
  totalRuns: number;
  isActive: boolean;
  createdAt: string;
}

// ─── Admin Dashboard Stats ───
export interface AdminStats {
  totalGamesThisMonth: number;
  activeGamesNow: number;
  totalPlayers: number;
  pendingHostRequests: number;
  pendingVoiceApprovals: number;
  rewardsFulfilled: number;
}
