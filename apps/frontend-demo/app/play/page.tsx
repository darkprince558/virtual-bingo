'use client'

import { useState, useEffect, useCallback, useMemo, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import { motion, AnimatePresence } from 'motion/react'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { BingoCard } from '@/components/BingoCard'
import { CurrentCallDisplay } from '@/components/CurrentCallDisplay'
import { CalledWordsFeed } from '@/components/CalledWordsFeed'
import { Leaderboard } from '@/components/Leaderboard'
import { AIHostPanel } from '@/components/AIHostPanel'
import { WinnerModal } from '@/components/WinnerModal'
import { BottomSheet } from '@/components/BottomSheet'
import { DecorativeBlobs } from '@/components/illustrations/DecorativeBlobs'
import { BingoCharacter } from '@/components/illustrations/BingoCharacter'
import { apiClient } from '@/lib/apiClient'
import { mapBingoPattern, mapClaimValidationReason, mapConnectionState, mapGameStatus } from '@/lib/uiMappers'
import { useGameEvents } from '@/hooks/useGameEvents'
import type { PlayerSnapshotResponse, CardResponse, ClaimSubmissionResponse, WinnerResponse, ClaimReadinessResponse } from '@/types/api'
import { BingoCellData } from '@/types/player'
import { Sparkles, Trophy, ChevronRight, Zap } from 'lucide-react'

function toPodiumPlacement(placement?: number): 1 | 2 | 3 {
  if (placement === 2) return 2
  if (placement === 3) return 3
  return 1
}

export default function PlayPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Loading...</div>}>
      <PlayContent />
    </Suspense>
  )
}

function PlayContent() {
  const searchParams = useSearchParams()
  const gameId = searchParams.get('gameId')
  const playerId = searchParams.get('playerId')

  const [snapshot, setSnapshot] = useState<PlayerSnapshotResponse | null>(null)
  const [cells, setCells] = useState<BingoCellData[]>([])
  const [showWinner, setShowWinner] = useState(false)
  const [isSheetOpen, setIsSheetOpen] = useState(false)
  const [sheetTab, setSheetTab] = useState<'leaderboard' | 'words' | 'ai'>('leaderboard')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [claimNotice, setClaimNotice] = useState<string | null>(null)
  const [confirmedWinner, setConfirmedWinner] = useState<WinnerResponse | null>(null)
  const [claimReadiness, setClaimReadiness] = useState<ClaimReadinessResponse | null>(null)
  const [isCheckingReadiness, setIsCheckingReadiness] = useState(false)

  const devAuth = useMemo(() => snapshot ? {
    devUserEmail: snapshot.player.email,
    devUserName: snapshot.player.displayName,
    devUserRole: 'player'
  } : {}, [snapshot])

  const cardToCells = useCallback((card: CardResponse): BingoCellData[] => (
    card.cells.map(c => ({
      id: c.id,
      word: c.word,
      isMarked: !!c.markedAt,
      isFreeSpace: c.isFreeSpace
    }))
  ), [])

  const fetchPlayerData = useCallback(async () => {
    if (!gameId || !playerId) {
      throw new Error('Missing game or player information. Join again from the lobby.')
    }

    const data = await apiClient<PlayerSnapshotResponse>(`/games/${gameId}/players/${playerId}/snapshot`)
    if (data.card) {
      return { snapshot: data, cells: cardToCells(data.card) }
    }

    const card = await apiClient<CardResponse>(`/games/${gameId}/players/me/card`, {
      method: 'POST',
      devUserEmail: data.player.email,
      devUserName: data.player.displayName,
      devUserRole: 'player'
    })
    return { snapshot: data, cells: cardToCells(card) }
  }, [cardToCells, gameId, playerId])

  const { lastEvent } = useGameEvents(snapshot ? gameId : null, devAuth)

  const refreshPlayerData = useCallback(async () => {
    const next = await fetchPlayerData()
    setSnapshot(next.snapshot)
    setCells(next.cells)
    setError(null)
    return next.snapshot
  }, [fetchPlayerData])

  useEffect(() => {
    let cancelled = false

    async function refresh() {
      try {
        const next = await fetchPlayerData()
        if (!cancelled) {
          setSnapshot(next.snapshot)
          setCells(next.cells)
          setError(null)
        }
      } catch (err) {
        console.error(err)
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load game')
        }
      }
    }

    void refresh()

    return () => {
      cancelled = true
    }
  }, [fetchPlayerData, lastEvent])

  useEffect(() => {
    if (!gameId || !snapshot) return

    let cancelled = false

    async function loadReadiness() {
      await Promise.resolve()
      if (cancelled) return
      setIsCheckingReadiness(true)
      try {
        const data = await apiClient<ClaimReadinessResponse>(`/games/${gameId}/players/me/claim-readiness`, {
          ...devAuth
        })
        if (!cancelled) {
          setClaimReadiness(data)
        }
      } catch (err) {
        console.error('Failed to load claim readiness:', err)
        if (!cancelled) {
          setClaimReadiness(null)
        }
      } finally {
        if (!cancelled) {
          setIsCheckingReadiness(false)
        }
      }
    }

    void loadReadiness()

    return () => {
      cancelled = true
    }
  }, [cells, devAuth, gameId, snapshot])

  useEffect(() => {
    if (!gameId || !snapshot) return

    let cancelled = false

    async function heartbeat() {
      try {
        const player = await apiClient<PlayerSnapshotResponse['player']>(`/games/${gameId}/players/me/heartbeat`, {
          method: 'POST',
          ...devAuth
        })
        if (!cancelled && player.reconnectNotice?.missedCalledWords?.length) {
          await refreshPlayerData()
        }
      } catch (err) {
        console.error('Player heartbeat failed:', err)
      }
    }

    void heartbeat()
    const interval = window.setInterval(() => {
      void heartbeat()
    }, 15000)

    return () => {
      cancelled = true
      window.clearInterval(interval)
    }
  }, [devAuth, gameId, refreshPlayerData, snapshot])

  const handleCellClick = async (id: string) => {
    if (!gameId || !snapshot) return
    const cell = cells.find(c => c.id === id)
    if (!cell) return

    // Optimistic update
    setCells(prev => prev.map(c => c.id === id ? { ...c, isMarked: !c.isMarked } : c))

    try {
      await apiClient(`/games/${gameId}/players/me/card/cells/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ marked: !cell.isMarked }),
        ...devAuth
      })
      await refreshPlayerData()
    } catch (err) {
      console.error('Failed to mark cell:', err)
      setClaimNotice('Could not update that square. Check the backend connection and try again.')
      // Revert on error
      setCells(prev => prev.map(c => c.id === id ? { ...c, isMarked: cell.isMarked } : c))
    }
  }

  const handleClaimBingo = async () => {
    if (!gameId || !snapshot || isSubmitting) return
    setIsSubmitting(true)
    setClaimNotice(null)
    try {
      if (!claimReadiness?.ready) {
        const reason = claimReadiness?.reason ? mapClaimValidationReason(claimReadiness.reason) : 'Backend will still validate this claim.'
        setClaimNotice(`Not quite ready yet: ${reason}`)
      }
      const result = await apiClient<ClaimSubmissionResponse>(`/games/${gameId}/claims`, {
        method: 'POST',
        body: JSON.stringify({ pattern: snapshot.winningPattern || 'single_line', playerId: snapshot.player.id }),
        ...devAuth
      })
      if (result.winner) {
        setConfirmedWinner(result.winner)
        setShowWinner(true)
        setClaimNotice('BINGO confirmed. Nice work.')
      } else {
        const reason = result.claim.validationResult?.reason
        setClaimNotice(`Claim submitted, but not confirmed yet${reason ? `: ${mapClaimValidationReason(reason)}` : '.'}`)
      }
      await refreshPlayerData()
    } catch (err) {
      console.error('Failed to claim BINGO:', err)
      setClaimNotice(err instanceof Error ? err.message : 'Failed to claim BINGO. Keep playing.')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (error) {
    return (
      <AppShell>
        <div className="flex-1 flex items-center justify-center p-6">
          <div className="max-w-md rounded-2xl p-6" style={{ background: '#FFFFFF', border: '1.5px solid #FECDD3', color: '#BE123C' }}>
            <p className="text-sm font-extrabold mb-2">Player game unavailable</p>
            <p className="text-sm font-semibold leading-relaxed">{error}</p>
          </div>
        </div>
      </AppShell>
    )
  }
  if (!snapshot) {
    return <AppShell><div className="flex-1 p-8">Loading...</div></AppShell>
  }

  const currentWord = snapshot.currentWord
  const markedCount = cells.filter(c => c.isMarked && !c.isFreeSpace).length
  const isCloseToBingo = markedCount >= 4
  const isVeryClose = markedCount >= 7
  const winnerPlacement = toPodiumPlacement(confirmedWinner?.placement)
  const missedWords = snapshot.reconnectNotice?.missedCalledWords || snapshot.player.reconnectNotice?.missedCalledWords || []
  const readinessPattern = claimReadiness?.bestPattern || claimReadiness?.readyPatterns?.[0] || snapshot.winningPattern

  return (
    <AppShell>
      <TopNav
        gameCode={snapshot.gameRun.code}
        playerName={snapshot.player.displayName}
        role="player"
        status={mapGameStatus(snapshot.status)}
        connectionState={mapConnectionState(snapshot.player.connectionState)}
      />

      <main className="flex-1 flex flex-col lg:flex-row overflow-y-auto lg:overflow-hidden relative">
        {/* Subtle background blobs */}
        <DecorativeBlobs variant="play" />

        {/* ── Main Game Area ── */}
        <div
          className="flex-1 p-4 sm:p-6 lg:p-8 flex flex-col gap-5 lg:overflow-y-auto relative z-10"
          style={{ overscrollBehavior: 'contain', scrollBehavior: 'smooth' }}
        >

          {/* Close-to-BINGO excitement banner */}
          <AnimatePresence>
            {isVeryClose && (
              <motion.div
                initial={{ opacity: 0, y: -20, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -20, scale: 0.95 }}
                className="flex items-center justify-center gap-3 py-3 rounded-lg relative overflow-hidden"
                style={{
                  background: 'linear-gradient(135deg, #FFF4F0 0%, #FFE4D9 100%)',
                  border: '2px solid #FFC5A8',
                }}
              >
                {/* Animated shimmer */}
                <div
                  className="absolute inset-0 animate-shimmer-glow pointer-events-none"
                  style={{
                    background: 'linear-gradient(90deg, transparent, rgba(255,90,31,0.08), transparent)',
                    backgroundSize: '200% 100%',
                  }}
                />
                <motion.div
                  animate={{ rotate: [0, 15, -15, 0] }}
                  transition={{ duration: 0.6, repeat: Infinity, repeatDelay: 1.5 }}
                >
                  <Zap className="w-5 h-5" style={{ color: '#FF5A1F' }} />
                </motion.div>
                <span className="text-sm font-extrabold relative z-10" style={{ color: '#E8440A' }}>
                  You&apos;re so close to BINGO!
                </span>
                <BingoCharacter mood="excited" size={36} />
              </motion.div>
            )}
          </AnimatePresence>

          {/* Current word call */}
          <CurrentCallDisplay
            word={currentWord?.word}
            aiMessage={snapshot.currentCallerAsset?.line || `Winning pattern: ${mapBingoPattern(snapshot.winningPattern)}`}
            callNumber={snapshot.calledWords.length + (currentWord ? 1 : 0)}
          />

          {missedWords.length > 0 && (
            <div className="rounded-lg p-4" style={{ background: '#FFFBEB', border: '1.5px solid #FDE68A', color: '#92400E' }}>
              <p className="text-sm font-extrabold">You reconnected after {missedWords.length} missed call{missedWords.length === 1 ? '' : 's'}.</p>
              <p className="text-xs font-semibold mt-1">
                Missed words: {missedWords.map(word => word.word).join(', ')}
              </p>
            </div>
          )}

          {/* Bingo Card */}
          <div className="flex-1 flex items-center justify-center">
            <BingoCard
              cells={cells}
              onCellClick={handleCellClick}
              currentWord={currentWord?.word}
            />
          </div>

          {/* Claim BINGO Button */}
          <div className="pb-2">
            <div className="max-w-2xl mx-auto mb-3 rounded-lg p-4" style={{ background: claimReadiness?.ready ? '#EDFAF5' : '#FFF4F0', border: `1.5px solid ${claimReadiness?.ready ? '#A8EBCC' : '#FFE4D9'}` }}>
              <p className="text-sm font-extrabold" style={{ color: claimReadiness?.ready ? '#116B3F' : '#C23208' }}>
                {isCheckingReadiness
                  ? 'Checking claim readiness...'
                  : claimReadiness?.ready
                    ? `Ready for ${mapBingoPattern(readinessPattern)}`
                    : `Not ready yet${claimReadiness?.reason ? `: ${mapClaimValidationReason(claimReadiness.reason)}` : ''}`}
              </p>
              {claimReadiness && !claimReadiness.ready && claimReadiness.missingCells.length > 0 && (
                <p className="text-xs font-semibold mt-1" style={{ color: '#B45309' }}>
                  Missing: {claimReadiness.missingCells.slice(0, 5).map(cell => cell.word).join(', ')}
                </p>
              )}
            </div>
            <motion.button
              id="claimBingoBtn"
              onClick={handleClaimBingo}
              disabled={isSubmitting}
              whileTap={{ scale: 0.97 }}
              whileHover={{ scale: 1.02, y: -2 }}
              animate={isCloseToBingo
                ? { boxShadow: ['0 6px 20px rgba(255,90,31,0.30)', '0 10px 30px rgba(255,90,31,0.55)', '0 6px 20px rgba(255,90,31,0.30)'] }
                : {}
              }
              transition={isCloseToBingo
                ? { duration: 1.6, repeat: Infinity, ease: 'easeInOut' }
                : {}
              }
              className="w-full max-w-2xl mx-auto flex items-center justify-center gap-3 py-4 sm:py-5 rounded-xl font-black text-base sm:text-lg transition-all relative overflow-hidden"
              style={{
                background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 100%)',
                color: '#FFFFFF',
                boxShadow: '0 6px 24px rgba(255, 90, 31, 0.35)',
                opacity: isSubmitting ? 0.7 : 1,
              }}
              aria-label="Claim bingo"
            >
              {/* Animated shimmer overlay */}
              {isCloseToBingo && (
                <div
                  className="absolute inset-0 animate-shimmer-glow pointer-events-none"
                  style={{
                    background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.15), transparent)',
                    backgroundSize: '200% 100%',
                  }}
                />
              )}
              {isCloseToBingo && (
                <motion.div
                  animate={{ rotate: [0, 10, -10, 0] }}
                  transition={{ duration: 0.5, repeat: Infinity, repeatDelay: 1 }}
                >
                  <Sparkles className="w-5 h-5" />
                </motion.div>
              )}
              <span className="relative z-10">
                {isSubmitting ? 'CHECKING CLAIM...' : 'CLAIM BINGO'}
              </span>
              {isCloseToBingo && <span className="text-sm font-bold opacity-80 relative z-10">- You&apos;re close!</span>}
            </motion.button>
            {claimNotice && (
              <p className="mt-3 text-center text-sm font-bold" style={{ color: claimNotice.includes('confirmed') ? '#116B3F' : '#B45309' }}>
                {claimNotice}
              </p>
            )}
          </div>
        </div>

        {/* Mobile Bottom Nav */}
        <div className="lg:hidden sticky bottom-0 left-0 right-0 bg-[#FAF8F5]/95 backdrop-blur-md border-t border-[#F0EDE8] p-3 flex gap-2 z-20">
          {(['leaderboard', 'words', 'ai'] as const).map((tab) => {
            const labels = {
              leaderboard: { label: 'Board', icon: <Trophy className="w-5 h-5 mb-0.5" /> },
              words: { label: 'Words', icon: <ChevronRight className="w-5 h-5 mb-0.5" /> },
              ai: { label: 'AI Host', icon: <Sparkles className="w-5 h-5 mb-0.5" /> },
            }
            return (
              <button
                key={tab}
                onClick={() => {
                  setSheetTab(tab)
                  setIsSheetOpen(true)
                }}
                className="flex-1 flex flex-col items-center justify-center py-2.5 rounded-2xl text-[11px] font-extrabold uppercase tracking-wide transition-transform active:scale-95"
                style={{ color: '#78716C', background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 12px rgba(0,0,0,0.05)' }}
              >
                {labels[tab].icon}
                {labels[tab].label}
              </button>
            )
          })}
        </div>

        {/* ── Sidebar (Desktop) ── */}
        <aside
          className="hidden lg:flex w-[320px] xl:w-[360px] shrink-0 flex-col relative z-10"
          style={{
            borderLeft: '1.5px solid rgba(240, 237, 232, 0.8)',
            background: 'rgba(255,255,255,0.85)',
            transform: 'translateZ(0)',
            willChange: 'transform',
          }}
        >
          {/* Leaderboard panel */}
          <div className="p-5 border-b border-[#F4F2EF]">
            <Leaderboard entries={snapshot.winners.map(w => ({ placement: toPodiumPlacement(w.placement), player: { id: w.playerId, name: w.playerId === snapshot.player.id ? snapshot.player.displayName : 'Player', state: 'Confirmed Winner', connectionState: 'Connected' }, wordsMatched: 4 }))} />
          </div>

          {/* Called words panel */}
          <div className="flex-1 p-5 overflow-hidden flex flex-col" style={{ minHeight: '180px', overscrollBehavior: 'contain' }}>
            <CalledWordsFeed words={snapshot.calledWords.map(w => ({ id: w.id, word: w.word, calledAt: w.calledAt }))} />
          </div>

          {/* AI Host panel */}
          <div className="shrink-0">
            <AIHostPanel message="Alice is 1 mark away from a diagonal bingo! The tension is building... will 'Leverage' be called next?" />
          </div>
        </aside>
      </main>

      <BottomSheet
        isOpen={isSheetOpen}
        onClose={() => setIsSheetOpen(false)}
        title={sheetTab === 'leaderboard' ? 'Leaderboard' : sheetTab === 'words' ? 'Called Words' : 'AI Host'}
      >
        {sheetTab === 'leaderboard' && (
          <Leaderboard entries={snapshot.winners.map(w => ({ placement: toPodiumPlacement(w.placement), player: { id: w.playerId, name: w.playerId === snapshot.player.id ? snapshot.player.displayName : 'Player', state: 'Confirmed Winner', connectionState: 'Connected' }, wordsMatched: 4 }))} />
        )}
        {sheetTab === 'words' && (
          <div className="h-full flex flex-col">
            <CalledWordsFeed words={snapshot.calledWords.map(w => ({ id: w.id, word: w.word, calledAt: w.calledAt }))} />
          </div>
        )}
        {sheetTab === 'ai' && (
          <AIHostPanel message="Alice is 1 mark away from a diagonal bingo! The tension is building... will 'Leverage' be called next?" />
        )}
      </BottomSheet>

      <WinnerModal
        isOpen={showWinner}
        winner={{ id: snapshot.player.id, name: snapshot.player.displayName, state: 'Confirmed Winner', connectionState: 'Connected' }}
        placement={winnerPlacement}
        pattern={mapBingoPattern(confirmedWinner?.pattern || snapshot.winningPattern)}
        onClose={() => setShowWinner(false)}
      />
    </AppShell>
  )
}
