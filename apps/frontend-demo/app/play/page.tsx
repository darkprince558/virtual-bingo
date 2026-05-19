'use client'

import { useState, useEffect, useCallback, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { motion, AnimatePresence } from 'motion/react'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { BingoCard } from '@/components/BingoCard'
import { CurrentCallDisplay } from '@/components/CurrentCallDisplay'
import { CalledWordsFeed } from '@/components/CalledWordsFeed'
import { Leaderboard } from '@/components/Leaderboard'
import { AIHostPanel } from '@/components/AIHostPanel'
import { WinnerModal } from '@/components/WinnerModal'
import { apiClient } from '@/lib/apiClient'
import { useGameEvents } from '@/hooks/useGameEvents'
import type { PlayerSnapshotResponse, CardResponse, CalledWordResponse, ClaimResponse, WinnerResponse } from '@/types/api'
import { BingoCellData } from '@/types/player'
import { Sparkles, Trophy, ChevronRight } from 'lucide-react'

// Extended word bank for mock "call next word"
const WORD_BANK = [
  'Action Item', 'Bandwidth', 'Circle Back', 'Deep Dive',
  'Ecosystem', 'Friction', 'Granular', 'Hard Stop',
  'Ideate', 'Leverage', 'Moving Forward', 'On My Radar',
  'Pain Point', 'Ping Me', 'Reach Out', 'Scalable',
  'Synergy', 'Touch Base', 'Value Add', 'Aligned',
]

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
  const [activeTab, setActiveTab] = useState<'leaderboard' | 'words' | 'ai'>('leaderboard')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const devAuth = snapshot ? {
    devUserEmail: snapshot.player.email,
    devUserName: snapshot.player.displayName,
    devUserRole: 'player'
  } : {}

  const loadData = useCallback(async () => {
    if (!gameId || !playerId) return
    try {
      // For initial load, we don't have devAuth yet, but we might need it.
      // We assume dev.go allows fetching if we just have X-Dev-User-ID or email.
      // But we can also rely on the URL params to fetch snapshot.
      // In dev mode, we pass some generic player headers or we must know the player's email.
      // Since we don't have the player email in URL, we will pass a generic one to fetch snapshot,
      // but actually GET /games/{gameID}/player-snapshot requires matching player.
      // Let's pass a placeholder that allows it, or fetch the player info first.
      // Assuming dev mode doesn't strictly check email if role=player for this simple test, or we fallback.
      const data = await apiClient<PlayerSnapshotResponse>(`/games/${gameId}/player-snapshot?playerId=${playerId}`)
      setSnapshot(data)

      if (data.card) {
        setCells(data.card.cells.map(c => ({
          id: c.id,
          word: c.word,
          isMarked: !!c.markedAt,
          isFreeSpace: c.isFreeSpace
        })))
      } else {
        // Generate card
        const card = await apiClient<CardResponse>(`/games/${gameId}/players/me/card`, {
          method: 'POST',
          devUserEmail: data.player.email,
          devUserName: data.player.displayName,
          devUserRole: 'player'
        })
        setCells(card.cells.map(c => ({
          id: c.id,
          word: c.word,
          isMarked: !!c.markedAt,
          isFreeSpace: c.isFreeSpace
        })))
      }
    } catch (err: any) {
      console.error(err)
      setError(err.message || 'Failed to load game')
    }
  }, [gameId, playerId])

  const { lastEvent } = useGameEvents(gameId, devAuth)

  useEffect(() => {
    if (lastEvent) {
      loadData()
    }
  }, [lastEvent, loadData])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadData()
  }, [loadData])

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
    } catch (err) {
      console.error('Failed to mark cell:', err)
      // Revert on error
      setCells(prev => prev.map(c => c.id === id ? { ...c, isMarked: cell.isMarked } : c))
    }
  }

  const handleClaimBingo = async () => {
    if (!gameId || !snapshot || isSubmitting) return
    setIsSubmitting(true)
    try {
      await apiClient(`/games/${gameId}/claims`, {
        method: 'POST',
        body: JSON.stringify({ pattern: 'Unknown', playerId: snapshot.player.id }),
        ...devAuth
      })
      setShowWinner(true)
    } catch (err) {
      console.error('Failed to claim BINGO:', err)
      alert('Failed to claim BINGO. Keep playing!')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (error) {
    return <AppShell><div className="flex-1 p-8 text-red-500">{error}</div></AppShell>
  }
  if (!snapshot) {
    return <AppShell><div className="flex-1 p-8">Loading...</div></AppShell>
  }

  const currentWord = snapshot.currentWord
  const markedCount = cells.filter(c => c.isMarked && !c.isFreeSpace).length
  const isCloseToBingo = markedCount >= 4

  return (
    <AppShell>
      <TopNav
        gameCode={snapshot.gameRun.code}
        playerName={snapshot.player.displayName}
        role="player"
        status={snapshot.status.charAt(0).toUpperCase() + snapshot.status.slice(1) as any}
        connectionState={snapshot.player.connectionState.charAt(0).toUpperCase() + snapshot.player.connectionState.slice(1) as any}
      />

      <main className="flex-1 flex flex-col lg:flex-row overflow-y-auto lg:overflow-hidden">

        {/* ── Main Game Area ── */}
        <div
          className="flex-1 p-4 sm:p-6 lg:p-8 flex flex-col gap-5 lg:overflow-y-auto"
          style={{ overscrollBehavior: 'contain', scrollBehavior: 'smooth' }}
        >

          {/* Current word call */}
          <CurrentCallDisplay
            word={currentWord?.word}
            aiMessage="Remember: 'Action Item' was used 4 times in last quarter's all-hands. Mark it if you hear it! 🎯"
          />

          {/* Bingo Card */}
          <div className="flex-1 flex items-center justify-center">
            <BingoCard cells={cells} onCellClick={handleCellClick} />
          </div>

          {/* Claim BINGO Button */}
          <div className="pb-2">
            <motion.button
              id="claimBingoBtn"
              onClick={handleClaimBingo}
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
              className="w-full max-w-2xl mx-auto flex items-center justify-center gap-3 py-4 sm:py-5 rounded-2xl font-black text-base sm:text-lg transition-all"
              style={{
                background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 100%)',
                color: '#FFFFFF',
                boxShadow: '0 6px 24px rgba(255, 90, 31, 0.35)',
              }}
              aria-label="Claim bingo"
            >
              {isCloseToBingo && (
                <motion.div
                  animate={{ rotate: [0, 10, -10, 0] }}
                  transition={{ duration: 0.5, repeat: Infinity, repeatDelay: 1 }}
                >
                  <Sparkles className="w-5 h-5" />
                </motion.div>
              )}
              CLAIM BINGO
              {isCloseToBingo && <span className="text-sm font-bold opacity-80">&mdash; You&apos;re close!</span>}
            </motion.button>
          </div>
        </div>

        {/* ── Sidebar (Desktop) ── */}
        <aside
          className="w-full lg:w-[320px] xl:w-[360px] shrink-0 flex flex-col"
          style={{
            borderLeft: '1.5px solid rgba(240, 237, 232, 0.8)',
            background: 'rgba(255,255,255,0.85)',
            transform: 'translateZ(0)',          /* own compositor layer — no scroll jank */
            willChange: 'transform',
          }}
        >
          {/* Mobile tab switcher */}
          <div
            className="flex lg:hidden px-4 pt-4 pb-0 gap-2 shrink-0"
            role="tablist"
            aria-label="Game info panels"
          >
            {(['leaderboard', 'words', 'ai'] as const).map((tab) => {
              const labels: Record<string, { label: string; icon: React.ReactNode }> = {
                leaderboard: { label: 'Board', icon: <Trophy className="w-3.5 h-3.5" /> },
                words: { label: 'Words', icon: <ChevronRight className="w-3.5 h-3.5" /> },
                ai: { label: 'AI Host', icon: <Sparkles className="w-3.5 h-3.5" /> },
              }
              const isActive = activeTab === tab
              return (
                <button
                  key={tab}
                  role="tab"
                  aria-selected={isActive}
                  onClick={() => setActiveTab(tab)}
                  className="flex-1 flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-extrabold transition-all"
                  style={{
                    background: isActive ? '#FFF4F0' : 'transparent',
                    color: isActive ? '#E8440A' : '#A8A29E',
                    border: isActive ? '1.5px solid #FFE4D9' : '1.5px solid transparent',
                  }}
                >
                  {labels[tab].icon}
                  {labels[tab].label}
                </button>
              )
            })}
          </div>

          {/* Leaderboard panel */}
          <div
            className={`p-5 ${activeTab !== 'leaderboard' ? 'hidden lg:block' : 'block'}`}
            style={{ borderBottom: '1px solid #F4F2EF' }}
          >
            <Leaderboard entries={snapshot.winners.map(w => ({ placement: w.placement, player: { id: w.playerId, name: w.playerId === snapshot.player.id ? snapshot.player.displayName : 'Player' }, wordsMatched: 4 })) as any} />
          </div>

          {/* Called words panel */}
          <div
            className={`flex-1 p-5 overflow-hidden flex flex-col ${activeTab !== 'words' ? 'hidden lg:flex' : 'flex'}`}
            style={{ minHeight: '180px', overscrollBehavior: 'contain' }}
          >
            <CalledWordsFeed words={snapshot.calledWords.map(w => ({ id: w.id, word: w.word, calledAt: w.calledAt }))} />
          </div>

          {/* AI Host panel */}
          <div className={`shrink-0 ${activeTab !== 'ai' ? 'hidden lg:block' : 'block'}`}>
            <AIHostPanel message="Alice is 1 mark away from a diagonal bingo! The tension is building... will 'Leverage' be called next?" />
          </div>
        </aside>
      </main>

      <WinnerModal
        isOpen={showWinner}
        winner={{ id: snapshot.player.id, name: snapshot.player.displayName, state: 'Confirmed Winner', connectionState: 'Connected' }}
        placement={1}
        pattern="Bingo!"
        onClose={() => setShowWinner(false)}
      />
    </AppShell>
  )
}
