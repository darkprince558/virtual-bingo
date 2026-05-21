'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { motion } from 'motion/react'
import {
  AlertCircle,
  ArrowRight,
  CheckCircle2,
  Clapperboard,
  ExternalLink,
  Loader2,
  MonitorPlay,
  Radio,
  Sparkles,
  UserRound,
} from 'lucide-react'
import { AppShell } from '@/components/AppShell'
import { apiClient } from '@/lib/apiClient'
import { displayBackendValue, mapGameStatus } from '@/lib/uiMappers'
import type {
  CalledWordResponse,
  CallerAssetResponse,
  CardCellResponse,
  CardResponse,
  ClaimSubmissionResponse,
  GameContentResponse,
  GameRunResponse,
  PlayerResponse,
} from '@/types/api'

const DEMO_CODE = 'LOCAL-DEMO'

const DEMO_PLAYERS = [
  { name: 'Alex Demo', email: 'alex@example.local', accent: '#FF5A1F' },
  { name: 'Sam Demo', email: 'sam@example.local', accent: '#7C5CFC' },
  { name: 'Taylor Demo', email: 'taylor@example.local', accent: '#22AA6A' },
]

type DemoStatus = 'idle' | 'loading' | 'ready' | 'error'

export default function DemoModePage() {
  const router = useRouter()
  const [status, setStatus] = useState<DemoStatus>('loading')
  const [game, setGame] = useState<GameRunResponse | null>(null)
  const [content, setContent] = useState<GameContentResponse | null>(null)
  const [assetsCount, setAssetsCount] = useState<number | null>(null)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isPreparing, setIsPreparing] = useState(false)
  const [isPrimingWinner, setIsPrimingWinner] = useState(false)
  const [joiningEmail, setJoiningEmail] = useState<string | null>(null)

  const gameId = game?.id

  const presenterLinks = useMemo(() => {
    if (!gameId) return []
    return [
      {
        label: 'Host Dashboard',
        href: '/host',
        icon: MonitorPlay,
        tone: '#FF5A1F',
      },
      {
        label: 'AI Review',
        href: `/host/review?gameId=${gameId}`,
        icon: Sparkles,
        tone: '#7C5CFC',
      },
      {
        label: 'Live Control',
        href: `/host/live?gameId=${gameId}`,
        icon: Radio,
        tone: '#22AA6A',
      },
      {
        label: 'Summary',
        href: `/summary?gameId=${gameId}`,
        icon: CheckCircle2,
        tone: '#F59E0B',
      },
    ]
  }, [gameId])

  const loadDemoGame = useCallback(async () => {
    await Promise.resolve()
    setStatus('loading')
    setError(null)
    try {
      const nextGame = await apiClient<GameRunResponse>(`/games/code/${DEMO_CODE}`)
      setGame(nextGame)
      try {
        const nextContent = await apiClient<GameContentResponse>(`/games/${nextGame.id}/content`, { devUserRole: 'host' })
        setContent(nextContent)
      } catch {
        setContent(null)
      }
      setStatus('ready')
    } catch (err) {
      setStatus('error')
      setError(err instanceof Error ? err.message : 'Demo game was not found')
    }
  }, [])

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadDemoGame()
    }, 0)
    return () => window.clearTimeout(timer)
  }, [loadDemoGame])

  async function prepareShowcase() {
    if (!gameId) return
    setIsPreparing(true)
    setMessage(null)
    setError(null)

    const notes: string[] = []
    try {
      let nextContent = content
      if (!nextContent) {
        nextContent = await apiClient<GameContentResponse>(`/games/${gameId}/content/prepare`, {
          method: 'POST',
          devUserRole: 'host',
        })
        notes.push('AI content prepared')
      } else {
        notes.push('AI content already prepared')
      }

      if (!nextContent.lockedAt) {
        nextContent = await apiClient<GameContentResponse>(`/games/${gameId}/content/lock`, {
          method: 'POST',
          devUserRole: 'host',
        })
        notes.push('word set locked')
      } else {
        notes.push('word set already locked')
      }

      const assets = await apiClient<CallerAssetResponse[]>(`/games/${gameId}/caller-assets/generate`, {
        method: 'POST',
        devUserRole: 'host',
      })
      setAssetsCount(assets.length)
      notes.push(`${assets.length} caller assets ready`)

      const nextGame = await apiClient<GameRunResponse>(`/games/${gameId}/lobby/open`, {
        method: 'POST',
        devUserRole: 'host',
      })
      setGame(nextGame)
      setContent(nextContent)
      notes.push('lobby open')
      setMessage(notes.join(' · '))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not prepare demo mode')
    } finally {
      setIsPreparing(false)
    }
  }

  async function primeWinnerMoment() {
    if (!gameId) return
    const player = DEMO_PLAYERS[0]
    setIsPrimingWinner(true)
    setMessage(null)
    setError(null)

    try {
      if (game?.status === 'finished' || game?.status === 'complete') {
        throw new Error('The demo game already has final results. Restart the backend with scripts/demo-backend.sh to reset LOCAL-DEMO.')
      }

      const devAuth = {
        devUserEmail: player.email,
        devUserName: player.name,
        devUserRole: 'player',
      }

      const joined = await apiClient<PlayerResponse>(`/games/${gameId}/players`, {
        method: 'POST',
        ...devAuth,
        body: JSON.stringify({ displayName: player.name, email: player.email }),
      })

      const card = await apiClient<CardResponse>(`/games/${gameId}/players/me/card`, {
        method: 'POST',
        ...devAuth,
      })

      const targetLine = pickBestSingleLine(card.cells)
      if (!targetLine.length) {
        throw new Error('Could not find a valid single-line pattern on the demo card.')
      }

      let activeGame = game
      if (activeGame?.status !== 'live') {
        activeGame = await apiClient<GameRunResponse>(`/games/${gameId}/start`, {
          method: 'POST',
          devUserRole: 'host',
        })
      }

      const called = await apiClient<CalledWordResponse[]>(`/games/${gameId}/calls`, {
        devUserRole: 'host',
      })
      const calledWords = new Set(called.map(word => normalizeWord(word.word)))
      const targetWords = targetLine.filter(cell => !cell.isFreeSpace).map(cell => normalizeWord(cell.word))

      let callsMade = 0
      while (targetWords.some(word => !calledWords.has(word)) && callsMade < 100) {
        const nextCall = await apiClient<CalledWordResponse>(`/games/${gameId}/calls`, {
          method: 'POST',
          devUserRole: 'host',
        })
        calledWords.add(normalizeWord(nextCall.word))
        callsMade += 1
      }

      const stillMissing = targetWords.filter(word => !calledWords.has(word))
      if (stillMissing.length > 0) {
        throw new Error(`Could not call every target word before the deck ran out. Missing: ${stillMissing.join(', ')}`)
      }

      for (const cell of targetLine) {
        if (cell.isFreeSpace) continue
        await apiClient(`/games/${gameId}/players/me/card/cells/${cell.id}`, {
          method: 'PATCH',
          ...devAuth,
          body: JSON.stringify({ marked: true }),
        })
      }

      const claim = await apiClient<ClaimSubmissionResponse>(`/games/${gameId}/claims`, {
        method: 'POST',
        ...devAuth,
        body: JSON.stringify({ pattern: 'single_line', playerId: joined.id }),
      })

      if (!claim.winner) {
        const reason = claim.claim.validationResult?.reason
        throw new Error(`Backend rejected the prepared claim${reason ? `: ${reason}` : '.'}`)
      }

      const nextGame = await apiClient<GameRunResponse>(`/games/${gameId}`, { devUserRole: 'host' })
      setGame(nextGame)
      setMessage(`${player.name} is now a real backend-confirmed winner. ${callsMade} word call${callsMade === 1 ? '' : 's'} were made to complete the line.`)
      router.push(`/play?gameId=${gameId}&playerId=${joined.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not prime the winner moment')
    } finally {
      setIsPrimingWinner(false)
    }
  }

  async function joinPlayer(player: typeof DEMO_PLAYERS[number], destination: 'lobby' | 'play') {
    if (!gameId) return
    setJoiningEmail(player.email)
    setError(null)
    try {
      if (destination === 'lobby') {
        router.push(`/lobby?code=${DEMO_CODE}&name=${encodeURIComponent(player.name)}&email=${encodeURIComponent(player.email)}`)
        return
      }

      const joined = await apiClient<PlayerResponse>(`/games/${gameId}/players`, {
        method: 'POST',
        devUserEmail: player.email,
        devUserName: player.name,
        devUserRole: 'player',
        body: JSON.stringify({ displayName: player.name, email: player.email }),
      })
      router.push(`/play?gameId=${gameId}&playerId=${joined.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : `Could not join as ${player.name}`)
      setJoiningEmail(null)
    }
  }

  async function startGame() {
    if (!gameId) return
    setError(null)
    try {
      const nextGame = await apiClient<GameRunResponse>(`/games/${gameId}/start`, {
        method: 'POST',
        devUserRole: 'host',
      })
      setGame(nextGame)
      router.push(`/host/live?gameId=${gameId}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not start the demo game')
    }
  }

  return (
    <AppShell>
      <main className="relative z-10 min-h-screen px-4 py-6 sm:px-8 sm:py-10">
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between"
          >
            <div>
              <div className="mb-3 inline-flex items-center gap-2 rounded-lg px-3 py-2 text-xs font-black uppercase tracking-widest" style={{ background: '#FFF4F0', color: '#E8440A' }}>
                <Clapperboard className="h-4 w-4" />
                Presenter Mode
              </div>
              <h1 className="text-3xl font-black tracking-tight sm:text-5xl" style={{ color: '#1C1917', lineHeight: 1.05 }}>
                Demo Control Room
              </h1>
              <p className="mt-3 max-w-2xl text-sm font-semibold sm:text-base" style={{ color: '#78716C' }}>
                Stable showcase links for the seeded Virtual Bingo game.
              </p>
            </div>

            <div className="flex flex-wrap gap-2">
              <Link href="/" className="inline-flex items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm font-extrabold" style={{ background: '#F4F2EF', color: '#57534E' }}>
                Home
              </Link>
              <button
                onClick={loadDemoGame}
                className="inline-flex items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm font-extrabold"
                style={{ background: '#1C1917', color: '#FFFFFF' }}
              >
                Refresh
              </button>
            </div>
          </motion.div>

          {status === 'error' && (
            <div className="rounded-lg p-4 text-sm font-bold" style={{ background: '#FFF1F2', border: '1.5px solid #FECDD3', color: '#BE123C' }}>
              {error || `Seed the ${DEMO_CODE} game before opening demo mode.`}
            </div>
          )}

          {error && status !== 'error' && (
            <div className="rounded-lg p-4 text-sm font-bold" style={{ background: '#FFF1F2', border: '1.5px solid #FECDD3', color: '#BE123C' }}>
              {error}
            </div>
          )}

          {message && (
            <div className="rounded-lg p-4 text-sm font-bold" style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC', color: '#116B3F' }}>
              {message}
            </div>
          )}

          <section className="grid grid-cols-1 gap-3 md:grid-cols-3">
            {[
              { label: 'BRD Pain', value: 'Manual cards, Teams calls, hand-checked winners' },
              { label: 'Demo Proof', value: 'Prepared game, live cards, backend validation' },
              { label: 'Roadmap Layer', value: 'Entra, Graph/Teams, rewards, Azure hosting' },
            ].map(item => (
              <div key={item.label} className="rounded-lg p-4" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>
                <p className="text-[10px] font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>{item.label}</p>
                <p className="mt-2 text-sm font-black leading-snug" style={{ color: '#1C1917' }}>{item.value}</p>
              </div>
            ))}
          </section>

          <section className="grid grid-cols-1 gap-4 lg:grid-cols-12">
            <div className="rounded-lg p-5 lg:col-span-4" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.05)' }}>
              <div className="mb-5 flex items-center justify-between gap-3">
                <div>
                  <p className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>Demo Game</p>
                  <h2 className="mt-1 text-xl font-black" style={{ color: '#1C1917' }}>{game?.name || 'Loading...'}</h2>
                </div>
                {status === 'loading' ? (
                  <Loader2 className="h-5 w-5 animate-spin" style={{ color: '#A8A29E' }} />
                ) : (
                  <span className="rounded-lg px-3 py-2 text-xs font-black" style={{ background: '#FFF4F0', color: '#E8440A' }}>{game?.code || DEMO_CODE}</span>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <Metric label="Status" value={game ? mapGameStatus(game.status) : 'Loading'} />
                <Metric label="Players" value={game ? String(game.allowedPlayerCount) : '-'} />
                <Metric label="AI Content" value={content?.lockedAt ? 'Locked' : content ? 'Draft' : 'Not ready'} />
                <Metric label="Caller Assets" value={assetsCount === null ? 'On demand' : String(assetsCount)} />
              </div>

              <button
                onClick={prepareShowcase}
                disabled={!gameId || isPreparing}
                className="mt-5 flex w-full items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm font-extrabold"
                style={{ background: '#FF5A1F', color: '#FFFFFF', opacity: !gameId || isPreparing ? 0.6 : 1 }}
              >
                {isPreparing ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                Prepare Showcase
              </button>

              <button
                onClick={startGame}
                disabled={!gameId || isPrimingWinner}
                className="mt-3 flex w-full items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm font-extrabold"
                style={{ background: '#1C1917', color: '#FFFFFF', opacity: !gameId || isPrimingWinner ? 0.6 : 1 }}
              >
                <Radio className="h-4 w-4" />
                Start Live Demo
              </button>

              <button
                onClick={primeWinnerMoment}
                disabled={!gameId || isPrimingWinner || isPreparing}
                className="mt-3 flex w-full items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm font-extrabold"
                style={{ background: '#F59E0B', color: '#FFFFFF', opacity: !gameId || isPrimingWinner || isPreparing ? 0.6 : 1 }}
              >
                {isPrimingWinner ? <Loader2 className="h-4 w-4 animate-spin" /> : <CheckCircle2 className="h-4 w-4" />}
                Prime Winner Moment
              </button>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:col-span-8">
              {presenterLinks.map(link => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="group rounded-lg p-5 transition-all hover:-translate-y-0.5"
                  style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.04)' }}
                >
                  <div className="flex items-center justify-between gap-4">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg" style={{ background: `${link.tone}18`, color: link.tone }}>
                        <link.icon className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-base font-black" style={{ color: '#1C1917' }}>{link.label}</p>
                        <p className="text-xs font-bold" style={{ color: '#A8A29E' }}>{displayLinkHint(link.href)}</p>
                      </div>
                    </div>
                    <ExternalLink className="h-4 w-4 transition-transform group-hover:translate-x-0.5" style={{ color: '#A8A29E' }} />
                  </div>
                </Link>
              ))}
            </div>
          </section>

          <section className="rounded-lg p-5" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.04)' }}>
            <div className="mb-4 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>Player Shortcuts</p>
                <h2 className="mt-1 text-xl font-black" style={{ color: '#1C1917' }}>Seeded Audience</h2>
              </div>
              <div className="inline-flex items-center gap-2 rounded-lg px-3 py-2 text-xs font-black" style={{ background: '#F5F2FF', color: '#6440E8' }}>
                <UserRound className="h-4 w-4" />
                Dev auth ready
              </div>
            </div>

            <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
              {DEMO_PLAYERS.map(player => (
                <div key={player.email} className="rounded-lg p-4" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}>
                  <div className="mb-4 flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg text-sm font-black" style={{ background: `${player.accent}18`, color: player.accent }}>
                      {player.name.split(' ').map(part => part[0]).join('')}
                    </div>
                    <div>
                      <p className="text-sm font-black" style={{ color: '#1C1917' }}>{player.name}</p>
                      <p className="text-xs font-bold" style={{ color: '#A8A29E' }}>{player.email}</p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => joinPlayer(player, 'lobby')}
                      disabled={!gameId || joiningEmail === player.email}
                      className="flex flex-1 items-center justify-center gap-1.5 rounded-lg px-3 py-2.5 text-xs font-extrabold"
                      style={{ background: '#F4F2EF', color: '#57534E', opacity: !gameId || joiningEmail === player.email ? 0.6 : 1 }}
                    >
                      Lobby
                    </button>
                    <button
                      onClick={() => joinPlayer(player, 'play')}
                      disabled={!gameId || joiningEmail === player.email}
                      className="flex flex-1 items-center justify-center gap-1.5 rounded-lg px-3 py-2.5 text-xs font-extrabold"
                      style={{ background: player.accent, color: '#FFFFFF', opacity: !gameId || joiningEmail === player.email ? 0.6 : 1 }}
                    >
                      {joiningEmail === player.email ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : 'Play'}
                      {joiningEmail !== player.email && <ArrowRight className="h-3.5 w-3.5" />}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="rounded-lg p-4" style={{ background: '#FFFBEB', border: '1.5px solid #FDE68A', color: '#92400E' }}>
            <div className="flex gap-3">
              <AlertCircle className="mt-0.5 h-5 w-5 shrink-0" />
              <p className="text-sm font-bold">
                Demo mode expects the local seed game. If this page cannot find it, run the backend migrations and seed `backend-go/internal/db/seeds/local_demo.sql`.
              </p>
            </div>
          </section>
        </div>
      </main>
    </AppShell>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg p-3" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}>
      <p className="text-[10px] font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>{label}</p>
      <p className="mt-1 truncate text-sm font-black" style={{ color: '#1C1917' }}>{displayBackendValue(value)}</p>
    </div>
  )
}

function displayLinkHint(href: string) {
  if (href.includes('/review')) return 'AI content and caller prep'
  if (href.includes('/live')) return 'Host calling screen'
  if (href.includes('/summary')) return 'Winners and recap'
  return 'Game setup and status'
}

function normalizeWord(word: string) {
  return word.trim().toLowerCase()
}

function pickBestSingleLine(cells: CardCellResponse[]) {
  const lines: CardCellResponse[][] = []

  for (let row = 0; row < 5; row += 1) {
    lines.push(cells.filter(cell => cell.rowIndex === row).sort((a, b) => a.colIndex - b.colIndex))
  }

  for (let col = 0; col < 5; col += 1) {
    lines.push(cells.filter(cell => cell.colIndex === col).sort((a, b) => a.rowIndex - b.rowIndex))
  }

  lines.push(cells.filter(cell => cell.rowIndex === cell.colIndex).sort((a, b) => a.rowIndex - b.rowIndex))
  lines.push(cells.filter(cell => cell.rowIndex + cell.colIndex === 4).sort((a, b) => a.rowIndex - b.rowIndex))

  return lines
    .filter(line => line.length === 5)
    .sort((a, b) => countPlayableCells(a) - countPlayableCells(b))[0] || []
}

function countPlayableCells(cells: CardCellResponse[]) {
  return cells.filter(cell => !cell.isFreeSpace).length
}
