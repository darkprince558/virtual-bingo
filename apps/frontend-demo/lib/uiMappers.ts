import type { ConnectionState, PlayerState } from '@/types/player'
import type { ActivityEvent, BingoClaim, BingoPattern, GameState } from '@/types/game'
import type { ActivityEventResponse } from '@/types/api'

function humanize(value: string) {
  return value
    .split(/[_\s-]+/)
    .filter(Boolean)
    .map(part => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join(' ')
}

export function mapGameStatus(status?: string | null): GameState {
  switch ((status || '').toLowerCase()) {
    case 'live':
      return 'Live'
    case 'paused':
      return 'Paused'
    case 'finished':
    case 'complete':
      return 'Finished'
    case 'cancelled':
      return 'Cancelled'
    case 'failed':
      return 'Failed'
    case 'invites_sent':
    case 'scheduled':
    case 'content_generating':
    case 'content_review':
      return 'Starting Soon'
    case 'lobby_open':
      return 'Lobby Open'
    case 'draft':
    default:
      return 'Waiting'
  }
}

export function mapPlayerState(state?: string | null): PlayerState {
  switch ((state || '').toLowerCase()) {
    case 'playing':
      return 'Playing'
    case 'claimed_bingo':
    case 'claimed bingo':
      return 'Claimed Bingo'
    case 'confirmed_winner':
    case 'confirmed winner':
      return 'Confirmed Winner'
    case 'rejected_claim':
    case 'rejected claim':
      return 'Rejected Claim'
    case 'disconnected':
      return 'Disconnected'
    case 'joined':
      return 'Joined'
    case 'waiting':
    default:
      return 'Waiting'
  }
}

export function mapConnectionState(state?: string | null): ConnectionState {
  switch ((state || '').toLowerCase()) {
    case 'online':
    case 'connected':
      return 'Connected'
    case 'reconnecting':
      return 'Reconnecting'
    case 'offline':
    case 'disconnected':
    default:
      return 'Disconnected'
  }
}

export function mapClaimStatus(status?: string | null): BingoClaim['status'] {
  switch ((status || '').toLowerCase()) {
    case 'valid':
      return 'Valid'
    case 'confirmed':
      return 'Confirmed'
    case 'invalid':
    case 'rejected':
      return 'Invalid'
    case 'pending':
    default:
      return 'Pending'
  }
}

export function mapBingoPattern(pattern?: string | null): BingoPattern {
  switch ((pattern || '').toLowerCase()) {
    case 'four_corners':
    case 'four corners':
      return 'Four Corners'
    case 'full_house':
    case 'full house':
      return 'Full House'
    case 'single_line':
    case 'line':
    default:
      return 'Line'
  }
}

export function displayBackendValue(value?: string | null) {
  return value ? humanize(value) : 'Unknown'
}

export function mapMarkingMode(mode?: string | null) {
  switch ((mode || '').toLowerCase()) {
    case 'auto_mark':
      return 'Auto Mark'
    case 'assist':
      return 'Assist'
    case 'manual':
    default:
      return 'Manual'
  }
}

export function mapContentStatus(status?: string | null) {
  switch ((status || '').toLowerCase()) {
    case 'generated':
      return 'Generated'
    case 'edited':
      return 'Edited'
    case 'locked':
      return 'Locked'
    case 'failed':
      return 'Failed'
    case 'draft':
    default:
      return 'Draft'
  }
}

export function mapClaimValidationReason(reason?: string | null) {
  switch ((reason || '').toLowerCase()) {
    case 'missing_called_words':
    case 'missing called words':
      return 'Missing Called Words'
    case 'unsupported_pattern':
      return 'Unsupported Pattern'
    case 'pattern_mismatch':
      return 'Wrong Winning Pattern'
    case 'game_not_live':
      return 'Game Not Live'
    default:
      return displayBackendValue(reason)
  }
}

function payloadString(payload: Record<string, unknown> | undefined, key: string) {
  const value = payload?.[key]
  return typeof value === 'string' ? value : undefined
}

function payloadNumber(payload: Record<string, unknown> | undefined, key: string) {
  const value = payload?.[key]
  return typeof value === 'number' ? value : undefined
}

export function mapActivityEvent(event: ActivityEventResponse): ActivityEvent {
  const payload = event.payload
  switch (event.type) {
    case 'game.created':
      return { id: event.id, label: 'Game Created', detail: `Game code ${payloadString(payload, 'code') || 'created'}.`, createdAt: event.createdAt, tone: 'neutral' }
    case 'game.started':
      return { id: event.id, label: 'Game Started', detail: 'The session is now live.', createdAt: event.createdAt, tone: 'success' }
    case 'game.paused':
      return { id: event.id, label: 'Game Paused', detail: 'The host paused the session.', createdAt: event.createdAt, tone: 'warning' }
    case 'game.resumed':
      return { id: event.id, label: 'Game Resumed', detail: 'The host resumed the session.', createdAt: event.createdAt, tone: 'success' }
    case 'game.finished':
      return { id: event.id, label: 'Game Finished', detail: displayBackendValue(payloadString(payload, 'reason') || 'finished'), createdAt: event.createdAt, tone: 'danger' }
    case 'content.generated':
      return { id: event.id, label: 'AI Content Ready', detail: 'Generated words are ready for review.', createdAt: event.createdAt, tone: 'success' }
    case 'content.edited':
      return { id: event.id, label: 'Content Edited', detail: 'Host edited the generated content.', createdAt: event.createdAt, tone: 'neutral' }
    case 'content.locked':
      return { id: event.id, label: 'Content Locked', detail: 'The word set and call deck were locked.', createdAt: event.createdAt, tone: 'success' }
    case 'call_deck.locked':
      return { id: event.id, label: 'Call Deck Locked', detail: `${payloadNumber(payload, 'words') || 'All'} words are ready to call.`, createdAt: event.createdAt, tone: 'success' }
    case 'word.called':
      return { id: event.id, label: 'Word Called', detail: payloadString(payload, 'word') ? `"${payloadString(payload, 'word')}" was called.` : 'A new word was called.', createdAt: event.createdAt, tone: 'neutral' }
    case 'player.joined':
      return { id: event.id, label: 'Player Joined', detail: `${payloadString(payload, 'displayName') || 'A player'} joined the game.`, createdAt: event.createdAt, tone: 'success' }
    case 'player.reconnected':
      return { id: event.id, label: 'Player Reconnected', detail: `${payloadString(payload, 'displayName') || 'A player'} reconnected.`, createdAt: event.createdAt, tone: 'success' }
    case 'player.disconnected':
      return { id: event.id, label: 'Player Disconnected', detail: 'A player missed the heartbeat window.', createdAt: event.createdAt, tone: 'warning' }
    case 'claim.submitted':
      return { id: event.id, label: 'Claim Submitted', detail: `Pattern: ${mapBingoPattern(payloadString(payload, 'pattern'))}.`, createdAt: event.createdAt, tone: 'warning' }
    case 'claim.validated':
      return { id: event.id, label: 'Claim Validated', detail: `Backend result: ${displayBackendValue(payloadString(payload, 'status'))}.`, createdAt: event.createdAt, tone: payloadString(payload, 'status') === 'confirmed' ? 'success' : 'danger' }
    case 'claim.acknowledged':
      return { id: event.id, label: 'Claim Acknowledged', detail: `Host acknowledged ${displayBackendValue(payloadString(payload, 'decision'))}${payloadString(payload, 'note') ? `: ${payloadString(payload, 'note')}` : '.'}`, createdAt: event.createdAt, tone: payloadString(payload, 'decision') === 'approve' ? 'success' : 'danger' }
    case 'winner.created':
      return { id: event.id, label: 'Winner Confirmed', detail: `Placement ${payloadNumber(payload, 'placement') || ''} recorded.`, createdAt: event.createdAt, tone: 'success' }
    case 'game.settings_updated':
      return { id: event.id, label: 'Settings Updated', detail: 'Game settings were changed.', createdAt: event.createdAt, tone: 'neutral' }
    case 'delivery.batch_created':
      return { id: event.id, label: 'Invites Prepared', detail: `${payloadNumber(payload, 'attempts') || 0} invite attempts created.`, createdAt: event.createdAt, tone: 'success' }
    default:
      return { id: event.id, label: displayBackendValue(event.type), detail: event.entityId ? `Entity: ${event.entityId}` : 'Backend activity recorded.', createdAt: event.createdAt, tone: 'neutral' }
  }
}
