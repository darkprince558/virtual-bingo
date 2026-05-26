'use client'

import { useState, useEffect, useCallback, Suspense, useRef } from 'react'
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
import { Activity, AlertTriangle, BarChart2, ExternalLink, Inbox, Server, Settings2, ShieldCheck, SlidersHorizontal, Users } from 'lucide-react'

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
  const [mobileTab, setMobileTab] = useState<'monitor' | 'claims' | 'details'>('monitor')
  const [manualControlsOpen, setManualControlsOpen] = useState(false)
  const [detailsOpen, setDetailsOpen] = useState(false)
  const audioPlayedRef = useRef(false)

  useEffect(() => {
    if (snapshot?.status?.toLowerCase() === 'finished' && !audioPlayedRef.current) {
      audioPlayedRef.current = true
      try {
        const msg = new SpeechSynthesisUtterance("Bingo! We have a winner!")
        msg.rate = 1.0;
        msg.pitch = 1.1;
        window.speechSynthesis.speak(msg)
      } catch (e) {
        console.error('Speech synthesis failed', e)
      }
    }
    if (snapshot?.status?.toLowerCase() !== 'finished') {
      audioPlayedRef.current = false
    }
  }, [snapshot?.status])

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
  // Leaderboard points are derived locally until the backend exposes scoring details.
  const mappedWinners = snapshot.winners.map(w => ({ placement: w.placement, player: { id: w.playerId, name: snapshot.players.find(p => p.id === w.playerId)?.displayName || 'Unknown', state: 'Confirmed Winner' as const, connectionState: 'Connected' as const }, wordsMatched: 4 }))
  const pendingClaims = mappedClaims.filter(claim => claim.status === 'Pending')
  const connectedPlayers = mappedPlayers.filter(player => player.connectionState === 'Connected').length
  const isControllerRunning = displayStatus === 'Live'
  const controllerMessage = isControllerRunning
    ? 'Server game controller is calling words and tracking winners.'
    : displayStatus === 'Paused'
      ? 'Automation is paused. Resume when the room is ready.'
      : displayStatus === 'Finished'
        ? 'Game is finished. Summary is ready.'
        : 'Controller is waiting for the host to start the run.'

  return (
    <AppShell>
      <TopNav gameId={snapshot.gameRun.id} gameCode={snapshot.gameRun.code} playerName="Host" role="host" status={displayStatus} />

      <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8" style={{ overscrollBehavior: 'contain', scrollBehavior: 'smooth' }}>

        {/* Page header */}
        <div className="mb-5 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] mb-1.5" style={{ color: '#A8A29E' }}>
              Live Monitor
            </p>
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl sm:text-3xl font-black tracking-tight" style={{ color: '#1C1917', letterSpacing: 0 }}>
                {snapshot.gameRun.name}
              </h1>
              <GameStatusBadge status={displayStatus} />
            </div>
            <p className="text-sm font-semibold mt-1" style={{ color: '#A8A29E' }}>
              {error || controllerMessage}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2 shrink-0">
            <Link
              href={`/host/live?gameId=${gameId}`}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-2 px-4 py-3 rounded-lg text-xs font-extrabold transition-all"
              style={{ background: '#EDFAF5', color: '#116B3F', border: '1.5px solid #A8EBCC' }}
            >
              <ExternalLink className="w-3.5 h-3.5" />
              Monitor View
            </Link>
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

        <div className="mb-5 grid grid-cols-2 gap-3 lg:grid-cols-4">
          {[
            { icon: Server, label: 'Controller', value: isControllerRunning ? 'Auto' : displayStatus, color: isControllerRunning ? '#116B3F' : '#B45309', bg: isControllerRunning ? '#EDFAF5' : '#FFFBEB' },
            { icon: Users, label: 'Connected', value: `${connectedPlayers}/${snapshot.playerCount}`, color: '#6440E8', bg: '#F5F2FF' },
            { icon: BarChart2, label: 'Called', value: snapshot.calledWords.length, color: '#116B3F', bg: '#EDFAF5' },
            { icon: Inbox, label: 'Claims', value: pendingClaims.length, color: pendingClaims.length > 0 ? '#BE123C' : '#78716C', bg: pendingClaims.length > 0 ? '#FFE4E6' : '#F4F2EF' },
          ].map(stat => (
            <div key={stat.label} className="rounded-lg p-3 flex items-center gap-3" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 12px rgba(0,0,0,0.04)' }}>
              <div className="h-9 w-9 shrink-0 rounded-md flex items-center justify-center" style={{ background: stat.bg }}>
                <stat.icon className="h-4 w-4" style={{ color: stat.color }} />
              </div>
              <div className="min-w-0">
                <p className="text-lg font-black leading-none truncate" style={{ color: '#1C1917' }}>{stat.value}</p>
                <p className="text-[10px] font-bold uppercase tracking-wider" style={{ color: '#A8A29E' }}>{stat.label}</p>
              </div>
            </div>
          ))}
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 mb-24 lg:mb-0">
          <div className={`col-span-1 lg:col-span-8 flex-col gap-5 ${mobileTab === 'monitor' ? 'flex' : 'hidden lg:flex'}`}>
            <section className="rounded-xl p-4 sm:p-5" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
              <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-3">
                  <div className="h-11 w-11 rounded-lg flex items-center justify-center" style={{ background: isControllerRunning ? '#EDFAF5' : '#FFFBEB' }}>
                    <ShieldCheck className="h-5 w-5" style={{ color: isControllerRunning ? '#116B3F' : '#B45309' }} />
                  </div>
                  <div>
                    <h2 className="text-base font-black" style={{ color: '#1C1917' }}>Server game controller</h2>
                    <p className="text-xs font-bold" style={{ color: '#78716C' }}>{controllerMessage}</p>
                  </div>
                </div>
                {pendingClaims.length > 0 && (
                  <div className="inline-flex items-center gap-2 rounded-lg px-3 py-2 text-xs font-extrabold" style={{ background: '#FFE4E6', color: '#BE123C' }}>
                    <AlertTriangle className="h-4 w-4" /> {pendingClaims.length} claim{pendingClaims.length === 1 ? '' : 's'} need review
                  </div>
                )}
              </div>

              <CurrentCallDisplay word={snapshot.currentWord?.word} audioUrl={snapshot.currentCallerAsset?.audioUrl} />

              <div className="mt-4 rounded-lg" style={{ border: '1.5px solid #F0EDE8' }}>
                <button
                  onClick={() => setManualControlsOpen(open => !open)}
                  className="flex w-full items-center justify-between px-3 py-2.5 text-left text-xs font-extrabold uppercase tracking-wider"
                  style={{ color: '#78716C' }}
                >
                  <span className="inline-flex items-center gap-2"><SlidersHorizontal className="h-4 w-4" /> Manual override</span>
                  <span style={{ color: manualControlsOpen ? '#E8440A' : '#A8A29E' }}>{manualControlsOpen ? 'Open' : 'Closed'}</span>
                </button>
                {manualControlsOpen && (
                  <div className="border-t p-3" style={{ borderColor: '#F0EDE8' }}>
                    <HostControls
                      status={displayStatus}
                      onStart={handleStart}
                      onPause={handlePause}
                      onEnd={handleEnd}
                      onNextWord={handleNextWord}
                      variant="embedded"
                    />
                  </div>
                )}
              </div>
            </section>

            <section className="rounded-xl" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
              <button
                onClick={() => setDetailsOpen(open => !open)}
                className="flex w-full items-center justify-between p-4 text-left"
              >
                <span className="inline-flex items-center gap-2 text-sm font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
                  <Activity className="h-4 w-4" /> Operations details
                </span>
                <span className="text-xs font-bold" style={{ color: '#78716C' }}>{detailsOpen ? 'Hide' : 'Show'}</span>
              </button>
              {detailsOpen && (
                <div className="grid grid-cols-1 gap-4 border-t p-4 lg:grid-cols-2" style={{ borderColor: '#F0EDE8' }}>
                  <div className="min-h-[220px]">
                    <CalledWordsFeed words={mappedCalledWords} />
                  </div>
                  <ActivityFeed events={activityEvents} />
                </div>
              )}
            </section>
          </div>

          <div className="col-span-1 lg:col-span-4 flex flex-col gap-5">
            <div className={mobileTab === 'claims' ? 'block' : 'hidden lg:block'} style={{ height: '420px' }}>
              <BingoClaimQueue
                claims={mappedClaims}
                onApprove={handleApprove}
                onReject={handleReject}
              />
            </div>

            <div className={`flex-col gap-5 ${mobileTab === 'details' ? 'flex' : 'hidden lg:flex'}`}>
              <div style={{ height: '320px' }}>
                <PlayerList players={mappedPlayers} totalConnected={snapshot.playerCount} />
              </div>

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
        {(['monitor', 'claims', 'details'] as const).map((tab) => {
          const labels = {
            monitor: { label: 'Monitor', icon: <Server className="w-5 h-5 mb-0.5" /> },
            claims: { label: 'Claims', icon: <Inbox className="w-5 h-5 mb-0.5" /> },
            details: { label: 'Details', icon: <Settings2 className="w-5 h-5 mb-0.5" /> },
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
              {tab === 'claims' && pendingClaims.length > 0 && (
                <span className="absolute top-2 right-6 w-2.5 h-2.5 rounded-full bg-[#E11D48] animate-pulse-ring" />
              )}
            </button>
          )
        })}
      </div>
    </AppShell>
  )
}
