'use client'

import { useState, useEffect, useMemo, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { motion } from 'motion/react'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { GameStatusBadge } from '@/components/GameStatusBadge'
import { DecorativeBlobs } from '@/components/illustrations/DecorativeBlobs'
import { BingoCharacter } from '@/components/illustrations/BingoCharacter'
import { Users, ArrowRight, Clock, Shield } from 'lucide-react'
import { apiClient } from '@/lib/apiClient'
import { mapGameStatus } from '@/lib/uiMappers'
import { useGameEvents } from '@/hooks/useGameEvents'
import type { GameRunResponse, PlayerResponse, PlayerSnapshotResponse } from '@/types/api'

function getInitials(name: string) {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

const AVATAR_COLORS = [
  { bg: '#FFF4F0', text: '#E8440A', border: '#FFE4D9' },
  { bg: '#F5F2FF', text: '#6440E8', border: '#D9CCFF' },
  { bg: '#EDFAF5', text: '#116B3F', border: '#A8EBCC' },
  { bg: '#FEF3C7', text: '#B45309', border: '#FDE68A' },
  { bg: '#FFF1F2', text: '#BE123C', border: '#FECDD3' },
]

export default function LobbyPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Loading...</div>}>
      <LobbyContent />
    </Suspense>
  )
}

function LobbyContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const code = searchParams.get('code')
  const name = searchParams.get('name') || 'Player'
  const emailParam = searchParams.get('email')

  const [game, setGame] = useState<GameRunResponse | null>(null)
  const [player, setPlayer] = useState<PlayerResponse | null>(null)
  const [joined, setJoined] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [connectedPlayers, setConnectedPlayers] = useState<PlayerResponse[]>([])

  useEffect(() => {
    async function init() {
      if (!code) {
        setError('No game code provided')
        return
      }
      try {
        // Find game by code
        const gameRun = await apiClient<GameRunResponse>(`/games/code/${code}`)
        setGame(gameRun)
      } catch (err: any) {
        setError(err.message || 'Failed to find game')
      }
    }
    init()
  }, [code])

  const handleJoin = async () => {
    if (!game) return
    try {
      const emailLocalPart = name.trim().split(/\s+/)[0]?.toLowerCase().replace(/[^a-z0-9._-]/g, '') || 'player'
      const email = emailParam || `${emailLocalPart}@example.local`
      const joinedPlayer = await apiClient<PlayerResponse>(`/games/${game.id}/players`, {
        method: 'POST',
        body: JSON.stringify({ displayName: name, email })
      })
      setPlayer(joinedPlayer)
      setConnectedPlayers([joinedPlayer])
      setJoined(true)
    } catch (err: any) {
      setError(err.message || 'Failed to join game')
    }
  }

  // Poll for game start if joined using events
  const devAuth = useMemo(() => player ? {
    devUserEmail: player.email,
    devUserName: player.displayName,
    devUserRole: 'player'
  } : {}, [player])
  const { lastEvent } = useGameEvents(joined && game ? game.id : null, devAuth)

  useEffect(() => {
    async function checkStatus() {
      if (!game || !player) return
      try {
        const snapshot = await apiClient<PlayerSnapshotResponse>(`/games/${game.id}/players/${player.id}/snapshot`, devAuth)
        setConnectedPlayers([snapshot.player])
        if (snapshot.gameRun.status === 'live') {
          router.push(`/play?gameId=${game.id}&playerId=${player.id}`)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to refresh lobby status')
      }
    }

    if (joined) {
      checkStatus()
    }
  }, [joined, game, player, lastEvent, router, devAuth])

  if (error) {
    return (
      <AppShell>
        <div className="flex-1 flex flex-col items-center justify-center p-8">
          <p className="text-red-500 mb-4">{error}</p>
          <Link href="/" className="px-4 py-2 bg-gray-100 rounded">Go Back</Link>
        </div>
      </AppShell>
    )
  }

  if (!game) {
    return (
      <AppShell>
        <div className="flex-1 flex items-center justify-center">Looking for game...</div>
      </AppShell>
    )
  }

  const VISIBLE_PLAYERS = connectedPlayers.slice(0, 10)
  const TOTAL_CONNECTED = connectedPlayers.length

  return (
    <AppShell>
      <TopNav gameId={game.id} gameCode={game.code} playerName={name} role="player" status={mapGameStatus(game.status)} />

      <main className="flex-1 flex items-center justify-center p-4 sm:p-8 relative">
        {/* Animated background blobs */}
        <DecorativeBlobs variant="lobby" />

        <div className="w-full max-w-2xl relative z-10">

          {/* Header with character */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center mb-8"
          >
            {/* Waiting character illustration */}
            <motion.div
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.2, type: 'spring' }}
              className="flex justify-center mb-4"
            >
              <BingoCharacter mood="waiting" size={80} />
            </motion.div>

            <div className="flex items-center justify-center gap-3 mb-4">
              <GameStatusBadge status={mapGameStatus(game.status)} />
            </div>
            <h1 className="text-4xl sm:text-5xl font-black tracking-tight mb-3" style={{ color: '#1C1917' }}>
              Game Lobby
            </h1>
            <p className="text-base font-semibold" style={{ color: '#78716C' }}>
              The host will start the game soon. Get ready!
            </p>
          </motion.div>

          {/* Game Code Card */}
          <motion.div
            initial={{ opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="rounded-xl p-6 sm:p-8 mb-5"
            style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.05)' }}
          >
            <div className="flex flex-col sm:flex-row items-center gap-6">
              {/* Code bubble */}
              <div className="flex flex-col items-center gap-2">
                <p className="text-[10px] font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>Game Code</p>
                <div
                  className="px-8 py-4 rounded-xl"
                  style={{ background: '#FFF4F0', border: '2px solid #FFE4D9' }}
                >
                  <span className="text-3xl font-black tracking-[0.2em]" style={{ color: '#FF5A1F', letterSpacing: '0.2em' }}>
                    {game.code}
                  </span>
                </div>
              </div>

              <div className="hidden sm:block h-16 w-px" style={{ background: '#F0EDE8' }} />

              {/* Stats */}
              <div className="flex gap-8 sm:gap-10">
                <div className="text-center">
                  <p className="text-3xl font-black" style={{ color: '#1C1917' }}>{TOTAL_CONNECTED}</p>
                  <p className="text-xs font-bold" style={{ color: '#A8A29E' }}>Players Joined</p>
                </div>
                <div className="text-center">
                  <div className="flex items-center gap-1.5 justify-center">
                    <Shield className="w-4 h-4" style={{ color: '#7C5CFC' }} />
                    <p className="text-sm font-extrabold" style={{ color: '#1C1917' }}>Host</p>
                  </div>
                  <p className="text-xs font-bold" style={{ color: '#A8A29E' }}>Host</p>
                </div>
              </div>
            </div>
          </motion.div>

          {/* Players grid */}
          <motion.div
            initial={{ opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="rounded-xl p-6 mb-5"
            style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.05)' }}
          >
            <div className="flex items-center gap-2 mb-5">
              <Users className="w-4 h-4" style={{ color: '#A8A29E' }} />
              <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
                Players in Lobby
              </span>
            </div>

            <div className="flex flex-wrap gap-3">
              {VISIBLE_PLAYERS.map((p: any, i) => {
                const color = AVATAR_COLORS[i % AVATAR_COLORS.length]
                return (
                  <motion.div
                    key={p.id}
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ delay: 0.25 + i * 0.05, type: 'spring', stiffness: 300, damping: 20 }}
                    className="flex flex-col items-center gap-1.5"
                    style={{ minWidth: '64px' }}
                  >
                    {/* Avatar bubble (floating animation staggered) */}
                    <motion.div
                      animate={{ y: [0, -5, 0] }}
                      transition={{ duration: 2.5 + i * 0.3, repeat: Infinity, ease: 'easeInOut', delay: i * 0.2 }}
                      className="w-12 h-12 rounded-xl flex items-center justify-center text-sm font-black"
                      style={{ background: color.bg, border: `2px solid ${color.border}`, color: color.text }}
                    >
                      {getInitials(p.displayName || p.name)}
                    </motion.div>
                    <span className="text-[10px] font-bold text-center leading-tight" style={{ color: '#78716C', maxWidth: '60px' }}>
                      {(p.displayName || p.name).split(' ')[0]}
                    </span>
                  </motion.div>
                )
              })}

              {/* "+N more" bubble */}
              {TOTAL_CONNECTED > VISIBLE_PLAYERS.length && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 0.5, type: 'spring' }}
                  className="flex flex-col items-center gap-1.5"
                  style={{ minWidth: '64px' }}
                >
                  <div
                    className="w-12 h-12 rounded-xl flex items-center justify-center text-sm font-black"
                    style={{ background: '#F4F2EF', border: '2px dashed #D6D3D1', color: '#78716C' }}
                  >
                    +{TOTAL_CONNECTED - VISIBLE_PLAYERS.length}
                  </div>
                  <span className="text-[10px] font-bold text-center" style={{ color: '#A8A29E' }}>more</span>
                </motion.div>
              )}
            </div>
          </motion.div>

          {/* Waiting status + CTA */}
          <motion.div
            initial={{ opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            {!joined ? (
              <div className="flex flex-col sm:flex-row gap-3">
                <motion.button
                  whileTap={{ scale: 0.97 }}
                  whileHover={{ scale: 1.02, y: -2 }}
                  onClick={handleJoin}
                  id="confirmJoinBtn"
                  className="flex-1 flex items-center justify-center gap-2 py-4 rounded-xl font-extrabold text-base transition-all"
                  style={{
                    background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)',
                    color: '#FFFFFF',
                    boxShadow: '0 6px 20px rgba(255,90,31,0.35)',
                  }}
                >
                  <Clock className="w-5 h-5" />
                  I&apos;m Ready - Waiting for Host
                </motion.button>
                <Link
                  href="/"
                  className="flex items-center justify-center px-6 py-4 rounded-xl font-bold text-sm transition-all"
                  style={{ background: '#F4F2EF', color: '#78716C' }}
                >
                  Leave
                </Link>
              </div>
            ) : (
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ type: 'spring', stiffness: 300, damping: 22 }}
                className="flex flex-col gap-3"
              >
                {/* Breathing waiting indicator */}
                <div
                  className="flex flex-col items-center justify-center gap-4 py-6 rounded-xl relative overflow-hidden"
                  style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC' }}
                >
                  {/* Breathing rings */}
                  <div className="relative w-12 h-12 flex items-center justify-center">
                    <motion.div
                      className="absolute rounded-full"
                      style={{ width: 48, height: 48, border: '2px solid #22AA6A', opacity: 0.3 }}
                      animate={{ scale: [1, 1.6, 1], opacity: [0.3, 0.05, 0.3] }}
                      transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
                    />
                    <motion.div
                      className="absolute rounded-full"
                      style={{ width: 48, height: 48, border: '2px solid #22AA6A', opacity: 0.2 }}
                      animate={{ scale: [1, 1.6, 1], opacity: [0.2, 0.05, 0.2] }}
                      transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut', delay: 1 }}
                    />
                    <motion.div
                      animate={{ scale: [1, 1.15, 1] }}
                      transition={{ duration: 3, repeat: Infinity }}
                      className="w-4 h-4 rounded-full"
                      style={{ background: '#22AA6A' }}
                    />
                  </div>
                  <span className="font-extrabold" style={{ color: '#116B3F' }}>
                    Ready! Waiting for the host to start…
                  </span>
                </div>
                <button
                  disabled
                  id="goToGameBtn"
                  className="flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-bold transition-all"
                  style={{
                    background: '#1C1917',
                    color: '#FFFFFF',
                    boxShadow: '0 4px 16px rgba(0,0,0,0.12)',
                    opacity: 0.7
                  }}
                >
                  Waiting for host to start...
                  <ArrowRight className="w-4 h-4" />
                </button>
              </motion.div>
            )}
          </motion.div>

        </div>
      </main>
    </AppShell>
  )
}
