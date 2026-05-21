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
import { mapActivityEvent, mapBingoPattern, mapClaimStatus, mapConnectionState, mapGameStatus, mapPlayerState } from '@/lib/uiMappers'
import { useGameEvents } from '@/hooks/useGameEvents'
import type { HostSnapshotResponse, CalledWordResponse, ActivityEventResponse, ClaimAcknowledgementResponse } from '@/types/api'
import { ActivityEvent } from '@/types/game'
import Link from 'next/link'
import { Users, BarChart2, ExternalLink, Settings2, Inbox } from 'lucide-react'

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
  const [isLoading, setIsLoading] = useState(Boolean(gameId))
  const [error, setError] = useState<string | null>(null)
  const [mobileTab, setMobileTab] = useState<'controls' | 'queue' | 'roster'>('controls')

  const loadSnapshot = useCallback(async () => {
    if (!gameId) {
      throw new Error('Choose a live game from the host dashboard or quick start a new session.')
    }
    return apiClient<HostSnapshotResponse>(`/games/${gameId}/host-snapshot`)
  }, [gameId])

  const loadActivity = useCallback(async () => {
    if (!gameId) return []
    const events = await apiClient<ActivityEventResponse[]>(`/games/${gameId}/activity?limit=100`, { devUserRole: 'host' })
    return events.map(mapActivityEvent)
  }, [gameId])

  const refreshAll = useCallback(async () => {
    const [data, events] = await Promise.all([loadSnapshot(), loadActivity()])
    setSnapshot(data)
    setActivityEvents(events)
    setError(null)
  }, [loadActivity, loadSnapshot])

  const { lastEvent } = useGameEvents(gameId, { devUserRole: 'host' })

  useEffect(() => {
    if (!gameId) {
      return
    }

    let cancelled = false

    async function refresh() {
      try {
        const [data, events] = await Promise.all([loadSnapshot(), loadActivity()])
        if (!cancelled) {
          setSnapshot(data)
          setActivityEvents(events)
          setError(null)
        }
      } catch (err) {
        console.error('Failed to fetch host snapshot:', err)
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load host snapshot')
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    void refresh()

    return () => {
      cancelled = true
    }
  }, [gameId, lastEvent, loadActivity, loadSnapshot])

  const refreshSnapshot = async () => {
    if (!gameId) return
    try {
      await refreshAll()
    } catch (err) {
      console.error('Failed to fetch host snapshot:', err)
      setError(err instanceof Error ? err.message : 'Failed to load host snapshot')
    }
  }

  const handleStart = async () => {
    if (!gameId) return
    try {
      const endpoint = snapshot?.status === 'paused' ? 'resume' : 'start'
      await apiClient(`/games/${gameId}/${endpoint}`, { method: 'POST' })
      await refreshSnapshot()
    } catch (err) {
      console.error(err)
      setError(err instanceof Error ? err.message : 'Could not start or resume the game')
    }
  }

  const handlePause = async () => {
    if (!gameId) return
    try {
      await apiClient(`/games/${gameId}/pause`, { method: 'POST' })
      await refreshSnapshot()
    } catch (err) {
      console.error(err)
      setError(err instanceof Error ? err.message : 'Could not pause the game')
    }
  }

  const handleEnd = async () => {
    if (!gameId) return
    try {
      await apiClient(`/games/${gameId}/finish`, { method: 'POST' })
      await refreshSnapshot()
    } catch (err) {
      console.error(err)
      setError(err instanceof Error ? err.message : 'Could not end the game')
    }
  }

  const handleNextWord = async () => {
    if (!gameId) return
    try {
      await apiClient<CalledWordResponse>(`/games/${gameId}/calls`, { method: 'POST' })
      await refreshSnapshot()
    } catch (err) {
      console.error(err)
      setError(err instanceof Error ? err.message : 'Could not call the next word')
    }
  }

  const acknowledgeClaim = async (id: string, decision: 'approve' | 'reject') => {
    if (!gameId) return
    try {
      await apiClient<ClaimAcknowledgementResponse>(`/games/${gameId}/claims/${id}/acknowledge`, {
        method: 'POST',
        body: JSON.stringify({
          decision,
          note: decision === 'approve' ? 'Host acknowledged backend-confirmed claim.' : 'Host acknowledged backend-rejected claim.',
        }),
        devUserRole: 'host'
      })
      await refreshSnapshot()
    } catch (err) {
      console.error(err)
      setError(err instanceof Error ? err.message : 'Could not acknowledge claim')
    }
  }

  const handleApprove = (id: string) => {
    void acknowledgeClaim(id, 'approve')
  }

  const handleReject = (id: string) => {
    void acknowledgeClaim(id, 'reject')
  }

  if (!gameId) {
    return (
      <AppShell>
        <div className="flex-1 flex items-center justify-center p-6">
          <div className="max-w-md rounded-xl p-6 text-center" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
            <p className="text-xs font-extrabold uppercase tracking-[0.2em] mb-2" style={{ color: '#A8A29E' }}>Live Game</p>
            <h1 className="text-2xl font-black mb-2" style={{ color: '#1C1917' }}>No game selected</h1>
            <p className="text-sm font-semibold mb-5" style={{ color: '#78716C' }}>Open a live run from the host dashboard, or quick start a new session.</p>
            <Link href="/host" className="inline-flex items-center justify-center px-5 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FFF4F0', color: '#E8440A', border: '1.5px solid #FFE4D9' }}>
              Back to Host Dashboard
            </Link>
          </div>
        </div>
      </AppShell>
    )
  }

  if (isLoading || !snapshot) {
    return (
      <AppShell>
        <div className="flex-1 flex items-center justify-center p-6">
          {error ? (
            <div className="max-w-md rounded-xl p-6" style={{ background: '#FFFFFF', border: '1.5px solid #FECDD3', color: '#BE123C' }}>
              <p className="text-sm font-extrabold mb-2">Host view unavailable</p>
              <p className="text-sm font-semibold leading-relaxed">{error}</p>
            </div>
          ) : 'Loading game...'}
        </div>
      </AppShell>
    )
  }

  // Maps backend data to the frontend types used by components
  const displayStatus = mapGameStatus(snapshot.status)
  const mappedCalledWords = snapshot.calledWords.map(w => ({ id: w.id, word: w.word, calledAt: w.calledAt }))
  const mappedClaims = snapshot.claims.map(c => ({ id: c.id, playerId: c.playerId, playerName: snapshot.players.find(p => p.id === c.playerId)?.displayName || 'Unknown', pattern: mapBingoPattern(c.pattern), status: mapClaimStatus(c.status), claimedAt: c.claimedAt }))
  const mappedPlayers = snapshot.players.map(p => ({ id: p.id, name: p.displayName, state: mapPlayerState(p.state), connectionState: mapConnectionState(p.connectionState) }))
  // The mock leaderboard expects entries with specific points. Using mock for leaderboard until points logic is fully mapped if needed.
  const mappedWinners = snapshot.winners.map(w => ({ placement: w.placement, player: { id: w.playerId, name: snapshot.players.find(p => p.id === w.playerId)?.displayName || 'Unknown', state: 'Confirmed Winner' as const, connectionState: 'Connected' as const }, wordsMatched: 4 }))

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
              {error || 'Run the session, review claims, and keep the winner order auditable.'}
            </p>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            {/* Quick stats */}
            <div className="hidden sm:flex items-center gap-4 px-5 py-3 rounded-lg"
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
              href={`/summary?gameId=${gameId}`}
              id="viewSummaryLink"
              className="flex items-center gap-2 px-4 py-3 rounded-lg text-xs font-extrabold transition-all"
              style={{ background: '#F4F2EF', color: '#44403C', border: '1.5px solid #E7E5E4' }}
            >
              <ExternalLink className="w-3.5 h-3.5" />
              Summary
            </Link>
          </div>
        </div>

        {/* Main grid */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 mb-20 lg:mb-0">

          {/* ─── Left Column: Controls + Status ─── */}
          <div className={`col-span-1 lg:col-span-5 flex-col gap-5 ${mobileTab === 'controls' ? 'flex' : 'hidden lg:flex'}`}>

            {/* Current word */}
            <CurrentCallDisplay word={snapshot.currentWord?.word} audioUrl={snapshot.currentCallerAsset?.audioUrl} />

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
              className="rounded-xl p-5 flex flex-col"
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
            <div className={mobileTab === 'queue' ? 'block' : 'hidden lg:block'} style={{ height: '380px' }}>
              <BingoClaimQueue
                claims={mappedClaims}
                onApprove={handleApprove}
                onReject={handleReject}
              />
            </div>

            {/* Players & Leaderboard (Roster Tab) */}
            <div className={`flex-col gap-5 ${mobileTab === 'roster' ? 'flex' : 'hidden lg:flex'}`}>
              {/* Players */}
              <div style={{ height: '320px' }}>
                <PlayerList players={mappedPlayers} totalConnected={snapshot.playerCount} />
              </div>

              {/* Leaderboard */}
              <div
                className="rounded-xl p-5"
                style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
              >
                <Leaderboard entries={mappedWinners.length > 0 ? mappedWinners : []} />
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Mobile Bottom Nav */}
      <div className="lg:hidden fixed bottom-0 left-0 right-0 bg-[#FAF8F5]/95 backdrop-blur-md border-t border-[#F0EDE8] p-3 flex gap-2 z-50 pb-safe">
        {(['controls', 'queue', 'roster'] as const).map((tab) => {
          const labels = {
            controls: { label: 'Controls', icon: <Settings2 className="w-5 h-5 mb-0.5" /> },
            queue: { label: 'Queue', icon: <Inbox className="w-5 h-5 mb-0.5" /> },
            roster: { label: 'Roster', icon: <Users className="w-5 h-5 mb-0.5" /> },
          }
          return (
            <button
              key={tab}
              onClick={() => setMobileTab(tab)}
              className="flex-1 flex flex-col items-center justify-center py-2.5 rounded-2xl text-[11px] font-extrabold uppercase tracking-wide transition-transform active:scale-95"
              style={{
                color: mobileTab === tab ? '#E8440A' : '#78716C',
                background: mobileTab === tab ? '#FFF4F0' : '#FFFFFF',
                border: mobileTab === tab ? '1.5px solid #FFE4D9' : '1.5px solid #F0EDE8',
                boxShadow: mobileTab === tab ? 'none' : '0 4px 12px rgba(0,0,0,0.05)'
              }}
            >
              {labels[tab].icon}
              {labels[tab].label}
              {tab === 'queue' && mappedClaims.filter(c => c.status === 'Pending').length > 0 && (
                <span className="absolute top-2 right-6 w-2.5 h-2.5 rounded-full bg-[#E11D48] animate-pulse-ring" />
              )}
            </button>
          )
        })}
      </div>
    </AppShell>
  )
}
