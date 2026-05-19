'use client'

import { useState, useEffect, useCallback, Suspense } from 'react'
import { motion } from 'motion/react'
import { useSearchParams } from 'next/navigation'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { HostControls } from '@/components/HostControls'
import { CurrentCallDisplay } from '@/components/CurrentCallDisplay'
import { BingoClaimQueue } from '@/components/BingoClaimQueue'
import { PlayerList } from '@/components/PlayerList'
import { CalledWordsFeed } from '@/components/CalledWordsFeed'
import { Leaderboard } from '@/components/Leaderboard'
import { ActivityFeed } from '@/components/ActivityFeed'
import { GameStatusBadge } from '@/components/GameStatusBadge'
import { apiClient } from '@/lib/apiClient'
import { useGameEvents } from '@/hooks/useGameEvents'
import type { HostSnapshotResponse, CalledWordResponse, ClaimResponse, GameRunResponse, PlayerResponse } from '@/types/api'
import { ActivityEvent } from '@/types/game'
import Link from 'next/link'
import { Users, Trophy, BarChart2, ExternalLink } from 'lucide-react'

// WORD_BANK is no longer needed since backend generates the words

export default function HostPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Loading...</div>}>
      <HostLiveContent />
    </Suspense>
  )
}

function HostLiveContent() {
  const searchParams = useSearchParams()
  const gameId = searchParams.get('gameId')

  const [snapshot, setSnapshot] = useState<HostSnapshotResponse | null>(null)
  const [activityEvents, setActivityEvents] = useState<ActivityEvent[]>([])
  const [isLoading, setIsLoading] = useState(true)

  const fetchSnapshot = useCallback(async () => {
    if (!gameId) return
    try {
      const data = await apiClient<HostSnapshotResponse>(`/games/${gameId}/host-snapshot`)
      setSnapshot(data)
    } catch (err) {
      console.error('Failed to fetch host snapshot:', err)
    } finally {
      setIsLoading(false)
    }
  }, [gameId])

  const { lastEvent } = useGameEvents(gameId, { devUserRole: 'host' })

  useEffect(() => {
    if (lastEvent) {
      fetchSnapshot()
    }
  }, [lastEvent, fetchSnapshot])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchSnapshot()
  }, [fetchSnapshot])

  const addActivity = (label: string, detail: string, tone: ActivityEvent['tone']) => {
    const newEvent: ActivityEvent = {
      id: `evt-${Date.now()}`,
      label,
      detail,
      createdAt: new Date().toISOString(),
      tone,
    }
    setActivityEvents(prev => [...prev, newEvent])
  }

  const handleStart = async () => {
    if (!gameId) return
    try {
      await apiClient(`/games/${gameId}/start`, { method: 'POST' })
      addActivity('Game started', 'Host moved the session to Live.', 'success')
      fetchSnapshot()
    } catch (err) {
      console.error(err)
    }
  }

  const handlePause = async () => {
    if (!gameId) return
    try {
      await apiClient(`/games/${gameId}/pause`, { method: 'POST' })
      addActivity('Game paused', 'Host paused the session.', 'warning')
      fetchSnapshot()
    } catch (err) {
      console.error(err)
    }
  }

  const handleEnd = async () => {
    if (!gameId) return
    try {
      await apiClient(`/games/${gameId}/finish`, { method: 'POST' })
      addActivity('Game ended', 'Host ended the session.', 'danger')
      fetchSnapshot()
    } catch (err) {
      console.error(err)
    }
  }

  const handleNextWord = async () => {
    if (!gameId) return
    try {
      const word = await apiClient<CalledWordResponse>(`/games/${gameId}/calls`, { method: 'POST' })
      addActivity('Word called', `"${word.word}" is now visible to all players.`, 'neutral')
      fetchSnapshot()
    } catch (err) {
      console.error(err)
    }
  }

  // NOTE: Approve/Reject are phase 4 items on backend, mock for now
  const handleApprove = (id: string) => {
    addActivity('Claim approved', `Claim confirmed as valid.`, 'success')
  }

  const handleReject = (id: string) => {
    addActivity('Claim rejected', `Claim was rejected.`, 'danger')
  }

  if (isLoading || !snapshot) {
    return (
      <AppShell>
        <div className="flex-1 flex items-center justify-center">Loading game...</div>
      </AppShell>
    )
  }

  // Maps backend data to the frontend types used by components
  const displayStatus = snapshot.status.charAt(0).toUpperCase() + snapshot.status.slice(1) as any
  const mappedCalledWords = snapshot.calledWords.map(w => ({ id: w.id, word: w.word, calledAt: w.calledAt }))
  const mappedClaims = snapshot.claims.map(c => ({ id: c.id, playerName: snapshot.players.find(p => p.id === c.playerId)?.displayName || 'Unknown', pattern: c.pattern, status: c.status === 'valid' ? 'Confirmed' : c.status === 'invalid' ? 'Invalid' : 'Pending' } as any))
  const mappedPlayers = snapshot.players.map(p => ({ id: p.id, name: p.displayName, state: p.state, connectionState: p.connectionState } as any))
  // The mock leaderboard expects entries with specific points. Using mock for leaderboard until points logic is fully mapped if needed.
  const mappedWinners = snapshot.winners.map(w => ({ placement: w.placement, player: { id: w.playerId, name: snapshot.players.find(p => p.id === w.playerId)?.displayName || 'Unknown' }, wordsMatched: 4 } as any))

  return (
    <AppShell>
      <TopNav gameCode={snapshot.gameRun.code} playerName="Host" role="host" status={displayStatus} />

      <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8" style={{ overscrollBehavior: 'contain', scrollBehavior: 'smooth' }}>

        {/* Page header */}
        <div className="mb-6 sm:mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
          <div>
            <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] mb-1.5" style={{ color: '#A8A29E' }}>
              Host Dashboard
            </p>
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl sm:text-3xl font-black tracking-tight" style={{ color: '#1C1917' }}>
                Game Control Center
              </h1>
              <GameStatusBadge status={displayStatus} />
            </div>
            <p className="text-sm font-semibold mt-1" style={{ color: '#A8A29E' }}>
              Run the session, review claims, and keep the winner order auditable.
            </p>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            {/* Quick stats */}
            <div className="hidden sm:flex items-center gap-4 px-5 py-3 rounded-2xl"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}
            >
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4" style={{ color: '#7C5CFC' }} />
                <span className="text-sm font-extrabold" style={{ color: '#1C1917' }}>{snapshot.playerCount}</span>
                <span className="text-xs font-semibold" style={{ color: '#A8A29E' }}>Players</span>
              </div>
              <div className="w-px h-4" style={{ background: '#F0EDE8' }} />
              <div className="flex items-center gap-2">
                <BarChart2 className="w-4 h-4" style={{ color: '#22AA6A' }} />
                <span className="text-sm font-extrabold" style={{ color: '#1C1917' }}>{snapshot.calledWords.length}</span>
                <span className="text-xs font-semibold" style={{ color: '#A8A29E' }}>Called</span>
              </div>
            </div>

            <Link
              href="/summary"
              id="viewSummaryLink"
              className="flex items-center gap-2 px-4 py-3 rounded-2xl text-xs font-extrabold transition-all"
              style={{ background: '#F4F2EF', color: '#44403C', border: '1.5px solid #E7E5E4' }}
            >
              <ExternalLink className="w-3.5 h-3.5" />
              Summary
            </Link>
          </div>
        </div>

        {/* Main grid */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">

          {/* ─── Left Column: Controls + Status ─── */}
          <div className="col-span-1 lg:col-span-5 flex flex-col gap-5">

            {/* Current word */}
            <CurrentCallDisplay word={snapshot.currentWord?.word} />

            {/* Controls */}
            <HostControls
              status={displayStatus}
              onStart={handleStart}
              onPause={handlePause}
              onEnd={handleEnd}
              onNextWord={handleNextWord}
            />

            {/* Called words feed */}
            <div
              className="rounded-3xl p-5 flex flex-col"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', minHeight: '220px', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <CalledWordsFeed words={mappedCalledWords} />
            </div>

            {/* Activity feed */}
            <ActivityFeed events={activityEvents} />
          </div>

          {/* ─── Right Column: Claims + Players + Leaderboard ─── */}
          <div className="col-span-1 lg:col-span-7 flex flex-col gap-5">

            {/* Claim Queue */}
            <div style={{ height: '380px' }}>
              <BingoClaimQueue
                claims={mappedClaims}
                onApprove={handleApprove}
                onReject={handleReject}
              />
            </div>

            {/* Players */}
            <div style={{ height: '320px' }}>
              <PlayerList players={mappedPlayers} totalConnected={snapshot.playerCount} />
            </div>

            {/* Leaderboard */}
            <div
              className="rounded-3xl p-5"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <Leaderboard entries={mappedWinners.length > 0 ? mappedWinners : []} />
            </div>
          </div>
        </div>
      </main>
    </AppShell>
  )
}
