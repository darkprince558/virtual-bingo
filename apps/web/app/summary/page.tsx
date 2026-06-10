'use client'

import { useEffect, useMemo, useState, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import { motion } from 'motion/react'
import Link from 'next/link'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { apiClient } from '@/lib/apiClient'
import { displayBackendValue, mapBingoPattern, mapGameStatus } from '@/lib/uiMappers'
import type { GameSummaryResponse } from '@/types/api'
import { Home, Download, Share2, Clock, Users, Hash, Trophy, ChevronLeft, PartyPopper, Medal } from 'lucide-react'

const PODIUM_CONFIG = {
  1: { emoji: <Trophy size={32} color="#F59E0B" />, bg: 'linear-gradient(135deg, #FBBF24, #F59E0B)', color: '#FFFFFF', glow: 'rgba(245,158,11,0.30)', height: 96, label: '1st Place' },
  2: { emoji: <Medal size={28} color="#A8A29E" />, bg: 'linear-gradient(135deg, #D6D3D1, #A8A29E)', color: '#FFFFFF', glow: 'rgba(168,162,158,0.20)', height: 72, label: '2nd Place' },
  3: { emoji: <Medal size={28} color="#C0003D" />, bg: 'linear-gradient(135deg, #FFA070, #C0003D)', color: '#FFFFFF', glow: 'rgba(232,0,45,0.20)', height: 56, label: '3rd Place' },
}

function getInitials(name: string) {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

function SummaryContent() {
  const searchParams = useSearchParams()
  const gameId = searchParams.get('gameId')
  const [summary, setSummary] = useState<GameSummaryResponse | null>(null)
  const [isLoading, setIsLoading] = useState(Boolean(gameId))
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!gameId) return
    let cancelled = false

    async function loadSummary() {
      try {
        const data = await apiClient<GameSummaryResponse>(`/games/${gameId}/summary`)
        if (!cancelled) {
          setSummary(data)
          setError(null)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load game summary')
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    void loadSummary()

    return () => {
      cancelled = true
    }
  }, [gameId])

  const winnerRows = useMemo(() => {
    if (!summary) return []
    return summary.winners
      .slice()
      .sort((a, b) => a.placement - b.placement)
      .map(winner => {
        const player = summary.players.find(p => p.id === winner.playerId)
        return {
          ...winner,
          playerName: player?.displayName || 'Unknown Player',
        }
      })
  }, [summary])

  if (!gameId) {
    return (
      <AppShell>
        <div className="flex-1 flex items-center justify-center p-6">
          <div className="max-w-md rounded-xl p-6 text-center" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>
            <Trophy className="w-9 h-9 mx-auto mb-3" style={{ color: '#F59E0B' }} />
            <h1 className="text-2xl font-black mb-2" style={{ color: '#1C1917' }}>No summary selected</h1>
            <p className="text-sm font-semibold mb-5" style={{ color: '#78716C' }}>Open a completed or live run from the host dashboard to view real results.</p>
            <Link href="/host" className="inline-flex items-center gap-2 px-5 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FFF0F3', color: '#C40026' }}>
              <ChevronLeft className="w-4 h-4" /> Host Dashboard
            </Link>
          </div>
        </div>
      </AppShell>
    )
  }

  if (isLoading || !summary) {
    return (
      <AppShell>
        <div className="flex-1 flex items-center justify-center p-6">
          {error ? (
            <div className="max-w-md rounded-xl p-6" style={{ background: '#FFFFFF', border: '1.5px solid #FECDD3', color: '#BE123C' }}>
              <p className="text-sm font-extrabold mb-2">Summary unavailable</p>
              <p className="text-sm font-semibold leading-relaxed">{error}</p>
            </div>
          ) : 'Loading summary...'}
        </div>
      </AppShell>
    )
  }

  const stats = [
    { icon: Users, label: 'Total Players', value: summary.playerCount, color: '#7C5CFC' },
    { icon: Hash, label: 'Words Called', value: summary.calledWordCount, color: '#E8002D' },
    { icon: Clock, label: 'Claims', value: summary.claims.length, color: '#22AA6A' },
  ]

  return (
    <AppShell>
      <TopNav gameId={summary.gameRun.id} gameCode={summary.gameRun.code} playerName="Host" role="host" status={mapGameStatus(summary.status)} />

      <main className="flex-1 overflow-y-auto p-4 sm:p-8">
        <div className="max-w-3xl mx-auto">
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }} className="text-center mb-10">
            <div className="flex justify-center mb-4 text-orange-500"><PartyPopper size={48} /></div>
            <p className="text-xs font-extrabold uppercase tracking-[0.2em] mb-2" style={{ color: '#E8002D' }}>
              {displayBackendValue(summary.status)}
            </p>
            <h1 className="text-4xl sm:text-5xl font-black tracking-tight mb-3" style={{ color: '#1C1917' }}>
              {summary.gameRun.name}
            </h1>
            <p className="text-base font-semibold" style={{ color: '#78716C' }}>
              Code {summary.gameRun.code} · {summary.currentWord ? `Current word: ${summary.currentWord.word}` : 'No current word'}
            </p>
          </motion.div>

          <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5, delay: 0.1 }} className="rounded-2xl p-6 sm:p-8 mb-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.05)' }}>
            <h2 className="text-sm font-extrabold uppercase tracking-widest mb-8 text-center" style={{ color: '#A8A29E' }}>
              Winners Podium
            </h2>

            {winnerRows.length === 0 ? (
              <div className="text-center py-10">
                <Trophy className="w-10 h-10 mx-auto mb-3" style={{ color: '#D6D3D1' }} />
                <p className="text-sm font-bold" style={{ color: '#A8A29E' }}>No winners confirmed yet.</p>
              </div>
            ) : (
              <>
                <div className="flex items-end justify-center gap-4 mb-8">
                  {[2, 1, 3].map(place => {
                    const row = winnerRows.find(winner => winner.placement === place)
                    if (!row) return null
                    const cfg = PODIUM_CONFIG[place as 1 | 2 | 3]
                    return (
                      <motion.div key={row.id} initial={{ opacity: 0, scaleY: 0 }} animate={{ opacity: 1, scaleY: 1 }} transition={{ type: 'spring', stiffness: 200 }} className="flex flex-col items-center gap-3" style={{ transformOrigin: 'bottom' }}>
                        <div className="w-14 h-14 rounded-lg flex items-center justify-center text-xl font-black text-white" style={{ background: cfg.bg, boxShadow: `0 6px 20px ${cfg.glow}` }}>
                          {getInitials(row.playerName)}
                        </div>
                        <p className="text-xs font-bold text-center max-w-[80px] truncate" style={{ color: '#57534E' }}>{row.playerName.split(' ')[0]}</p>
                        <div className="w-20 rounded-t-2xl flex flex-col items-center justify-end pb-3" style={{ height: `${cfg.height}px`, background: '#F4F2EF' }}>
                          <span className="text-2xl">{cfg.emoji}</span>
                          <span className="text-[10px] font-black" style={{ color: '#78716C' }}>{place}{place === 1 ? 'st' : place === 2 ? 'nd' : 'rd'}</span>
                        </div>
                      </motion.div>
                    )
                  })}
                </div>

                <div className="space-y-2.5">
                  {winnerRows.map(row => {
                    const cfg = PODIUM_CONFIG[row.placement as 1 | 2 | 3] || PODIUM_CONFIG[3]
                    return (
                      <div key={row.id} className="flex items-center gap-4 px-4 py-3 rounded-md" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}>
                        <span className="text-xl">{cfg.emoji}</span>
                        <div className="w-9 h-9 rounded-md flex items-center justify-center text-xs font-black text-white shrink-0" style={{ background: cfg.bg }}>
                          {getInitials(row.playerName)}
                        </div>
                        <div className="flex-1">
                          <p className="text-sm font-bold" style={{ color: '#1C1917' }}>{row.playerName}</p>
                          <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>{mapBingoPattern(row.pattern)}</p>
                        </div>
                        <span className="text-xs font-extrabold px-3 py-1 rounded-full" style={{ background: '#F4F2EF', color: '#57534E' }}>
                          {cfg.label}
                        </span>
                      </div>
                    )
                  })}
                </div>
              </>
            )}
          </motion.div>

          <div className="grid grid-cols-3 gap-4 mb-6">
            {stats.map(stat => (
              <div key={stat.label} className="rounded-xl p-4 sm:p-5 flex flex-col items-center gap-2" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 12px rgba(0,0,0,0.04)' }}>
                <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ background: `${stat.color}15` }}>
                  <stat.icon className="w-5 h-5" style={{ color: stat.color }} />
                </div>
                <p className="text-2xl font-black" style={{ color: '#1C1917' }}>{stat.value}</p>
                <p className="text-[10px] font-bold text-center uppercase tracking-wide" style={{ color: '#A8A29E' }}>{stat.label}</p>
              </div>
            ))}
          </div>

          <div className="rounded-lg p-5 mb-6 flex items-center gap-4" style={{ background: '#F5F2FF', border: '1.5px solid #D9CCFF' }}>
            <div className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0" style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)', boxShadow: '0 4px 12px rgba(124,92,252,0.25)' }}>
              <Share2 className="w-5 h-5 text-white" />
            </div>
            <div className="flex-1">
              <p className="text-sm font-extrabold" style={{ color: '#4F30C2' }}>Prize Notification</p>
              <p className="text-xs font-semibold" style={{ color: '#7C5CFC' }}>Winner notifications via Microsoft integrations are still future functionality.</p>
            </div>
            <span className="text-[10px] font-extrabold px-2.5 py-1 rounded-full uppercase" style={{ background: '#EDE5FF', color: '#6440E8' }}>Soon</span>
          </div>

          <div className="flex flex-col sm:flex-row gap-3">
            <Link href="/host" id="returnHomeBtn" className="flex-1 flex items-center justify-center gap-2 py-4 rounded-xl font-extrabold text-base transition-all" style={{ background: 'linear-gradient(135deg, #C0003D, #E8002D)', color: '#FFFFFF', boxShadow: '0 6px 20px rgba(232,0,45,0.30)' }}>
              <Home className="w-5 h-5" /> Host Dashboard
            </Link>
            <button id="exportResultsBtn" className="flex items-center justify-center gap-2 px-6 py-4 rounded-xl font-bold text-sm transition-all" style={{ background: '#F4F2EF', color: '#78716C', border: '1.5px solid #E7E5E4' }} title="Export coming with backend integration">
              <Download className="w-4 h-4" /> Export Results
            </button>
          </div>
        </div>
      </main>
    </AppShell>
  )
}

export default function SummaryPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Loading...</div>}>
      <SummaryContent />
    </Suspense>
  )
}
