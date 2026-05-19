'use client'

import { motion } from 'motion/react'
import Link from 'next/link'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { mockGame, mockLeaderboard } from '@/lib/mockGameData'
import { Home, Download, Share2, Clock, Users, Hash } from 'lucide-react'

const WINNERS = mockLeaderboard.slice(0, 3)

const PODIUM_CONFIG = {
  1: { emoji: '🏆', bg: 'linear-gradient(135deg, #FBBF24, #F59E0B)', color: '#FFFFFF', glow: 'rgba(245,158,11,0.30)', height: 96, label: '1st Place' },
  2: { emoji: '🥈', bg: 'linear-gradient(135deg, #D6D3D1, #A8A29E)', color: '#FFFFFF', glow: 'rgba(168,162,158,0.20)', height: 72, label: '2nd Place' },
  3: { emoji: '🥉', bg: 'linear-gradient(135deg, #FFA070, #FF7A42)', color: '#FFFFFF', glow: 'rgba(255,90,31,0.20)', height: 56, label: '3rd Place' },
}

const GAME_STATS = [
  { icon: Users, label: 'Total Players', value: mockGame.connectedPlayers, color: '#7C5CFC' },
  { icon: Hash, label: 'Words Called', value: mockGame.calledWords.length + 1, color: '#FF5A1F' },
  { icon: Clock, label: 'Duration', value: '28 min', color: '#22AA6A' },
]

function getInitials(name: string) {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

export default function SummaryPage() {
  return (
    <AppShell>
      <TopNav gameCode={mockGame.code} playerName={mockGame.hostName} role="host" />

      <main className="flex-1 overflow-y-auto p-4 sm:p-8">
        <div className="max-w-3xl mx-auto">

          {/* Header */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center mb-10"
          >
            <div className="text-5xl mb-4 select-none">🎉</div>
            <p className="text-xs font-extrabold uppercase tracking-[0.2em] mb-2" style={{ color: '#FF5A1F' }}>
              Game Complete
            </p>
            <h1 className="text-4xl sm:text-5xl font-black tracking-tight mb-3" style={{ color: '#1C1917' }}>
              Final Results
            </h1>
            <p className="text-base font-semibold" style={{ color: '#78716C' }}>
              Thanks for playing, {mockGame.hostName}! Great session.
            </p>
          </motion.div>

          {/* Podium */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="rounded-3xl p-6 sm:p-8 mb-6"
            style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.05)' }}
          >
            <h2 className="text-sm font-extrabold uppercase tracking-widest mb-8 text-center" style={{ color: '#A8A29E' }}>
              🏆 Winners Podium
            </h2>

            {/* Podium bars (visual) */}
            <div className="flex items-end justify-center gap-4 mb-8">
              {/* 2nd place */}
              {WINNERS[1] && (
                <motion.div
                  initial={{ opacity: 0, scaleY: 0 }}
                  animate={{ opacity: 1, scaleY: 1 }}
                  transition={{ delay: 0.3, type: 'spring', stiffness: 200 }}
                  className="flex flex-col items-center gap-3"
                  style={{ transformOrigin: 'bottom' }}
                >
                  <div
                    className="w-14 h-14 rounded-2xl flex items-center justify-center text-xl font-black text-white"
                    style={{ background: PODIUM_CONFIG[2].bg, boxShadow: `0 6px 20px ${PODIUM_CONFIG[2].glow}` }}
                  >
                    {getInitials(WINNERS[1].player.name)}
                  </div>
                  <p className="text-xs font-bold text-center max-w-[80px] truncate" style={{ color: '#57534E' }}>
                    {WINNERS[1].player.name.split(' ')[0]}
                  </p>
                  <div
                    className="w-20 rounded-t-2xl flex flex-col items-center justify-end pb-3"
                    style={{ height: `${PODIUM_CONFIG[2].height}px`, background: 'linear-gradient(135deg, #F4F2EF, #E7E5E4)' }}
                  >
                    <span className="text-2xl">🥈</span>
                    <span className="text-[10px] font-black" style={{ color: '#78716C' }}>2nd</span>
                  </div>
                </motion.div>
              )}

              {/* 1st place */}
              {WINNERS[0] && (
                <motion.div
                  initial={{ opacity: 0, scaleY: 0 }}
                  animate={{ opacity: 1, scaleY: 1 }}
                  transition={{ delay: 0.15, type: 'spring', stiffness: 200 }}
                  className="flex flex-col items-center gap-3"
                  style={{ transformOrigin: 'bottom' }}
                >
                  <motion.div
                    animate={{ y: [0, -6, 0] }}
                    transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
                    className="w-16 h-16 rounded-2xl flex items-center justify-center text-xl font-black text-white"
                    style={{
                      background: PODIUM_CONFIG[1].bg,
                      boxShadow: `0 8px 24px ${PODIUM_CONFIG[1].glow}`,
                    }}
                  >
                    {getInitials(WINNERS[0].player.name)}
                  </motion.div>
                  <p className="text-sm font-extrabold text-center max-w-[80px] truncate" style={{ color: '#1C1917' }}>
                    {WINNERS[0].player.name.split(' ')[0]}
                  </p>
                  <div
                    className="w-20 rounded-t-2xl flex flex-col items-center justify-end pb-3"
                    style={{
                      height: `${PODIUM_CONFIG[1].height}px`,
                      background: 'linear-gradient(135deg, #FEF3C7, #FDE68A)',
                    }}
                  >
                    <span className="text-3xl">🏆</span>
                    <span className="text-[10px] font-black" style={{ color: '#B45309' }}>1st</span>
                  </div>
                </motion.div>
              )}

              {/* 3rd place */}
              {WINNERS[2] && (
                <motion.div
                  initial={{ opacity: 0, scaleY: 0 }}
                  animate={{ opacity: 1, scaleY: 1 }}
                  transition={{ delay: 0.45, type: 'spring', stiffness: 200 }}
                  className="flex flex-col items-center gap-3"
                  style={{ transformOrigin: 'bottom' }}
                >
                  <div
                    className="w-14 h-14 rounded-2xl flex items-center justify-center text-xl font-black text-white"
                    style={{ background: PODIUM_CONFIG[3].bg, boxShadow: `0 6px 20px ${PODIUM_CONFIG[3].glow}` }}
                  >
                    {getInitials(WINNERS[2].player.name)}
                  </div>
                  <p className="text-xs font-bold text-center max-w-[80px] truncate" style={{ color: '#57534E' }}>
                    {WINNERS[2].player.name.split(' ')[0]}
                  </p>
                  <div
                    className="w-20 rounded-t-2xl flex flex-col items-center justify-end pb-3"
                    style={{ height: `${PODIUM_CONFIG[3].height}px`, background: 'linear-gradient(135deg, #FFF4F0, #FFE4D9)' }}
                  >
                    <span className="text-2xl">🥉</span>
                    <span className="text-[10px] font-black" style={{ color: '#C23208' }}>3rd</span>
                  </div>
                </motion.div>
              )}
            </div>

            {/* Full placement list */}
            <div className="space-y-2.5">
              {WINNERS.map((entry, i) => {
                const cfg = PODIUM_CONFIG[(entry.placement as 1 | 2 | 3)]
                return (
                  <motion.div
                    key={entry.player.id}
                    initial={{ opacity: 0, x: -16 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: 0.5 + i * 0.1 }}
                    className="flex items-center gap-4 px-4 py-3 rounded-2xl"
                    style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                  >
                    <span className="text-xl">{cfg.emoji}</span>
                    <div
                      className="w-9 h-9 rounded-xl flex items-center justify-center text-xs font-black text-white shrink-0"
                      style={{ background: cfg.bg }}
                    >
                      {getInitials(entry.player.name)}
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-bold" style={{ color: '#1C1917' }}>{entry.player.name}</p>
                      <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>{entry.wordsMatched} words matched</p>
                    </div>
                    <span className="text-xs font-extrabold px-3 py-1 rounded-full"
                      style={{ background: '#F4F2EF', color: '#57534E' }}>
                      {cfg.label}
                    </span>
                  </motion.div>
                )
              })}
            </div>
          </motion.div>

          {/* Game stats */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.35, duration: 0.5 }}
            className="grid grid-cols-3 gap-4 mb-6"
          >
            {GAME_STATS.map((stat, i) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4 + i * 0.08 }}
                className="rounded-3xl p-4 sm:p-5 flex flex-col items-center gap-2"
                style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 12px rgba(0,0,0,0.04)' }}
              >
                <div
                  className="w-10 h-10 rounded-2xl flex items-center justify-center"
                  style={{ background: `${stat.color}15` }}
                >
                  <stat.icon className="w-5 h-5" style={{ color: stat.color }} />
                </div>
                <p className="text-2xl font-black" style={{ color: '#1C1917' }}>{stat.value}</p>
                <p className="text-[10px] font-bold text-center uppercase tracking-wide" style={{ color: '#A8A29E' }}>{stat.label}</p>
              </motion.div>
            ))}
          </motion.div>

          {/* Prize & export placeholders */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="rounded-3xl p-5 mb-6 flex items-center gap-4"
            style={{ background: '#F5F2FF', border: '1.5px solid #D9CCFF' }}
          >
            <div
              className="w-10 h-10 rounded-2xl flex items-center justify-center shrink-0"
              style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)', boxShadow: '0 4px 12px rgba(124,92,252,0.25)' }}
            >
              <Share2 className="w-5 h-5 text-white" />
            </div>
            <div className="flex-1">
              <p className="text-sm font-extrabold" style={{ color: '#4F30C2' }}>Prize Notification</p>
              <p className="text-xs font-semibold" style={{ color: '#7C5CFC' }}>Winner notifications via email — coming with Microsoft Entra integration.</p>
            </div>
            <span className="text-[10px] font-extrabold px-2.5 py-1 rounded-full uppercase" style={{ background: '#EDE5FF', color: '#6440E8' }}>
              Soon
            </span>
          </motion.div>

          {/* Actions */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.55 }}
            className="flex flex-col sm:flex-row gap-3"
          >
            <Link
              href="/"
              id="returnHomeBtn"
              className="flex-1 flex items-center justify-center gap-2 py-4 rounded-2xl font-extrabold text-base transition-all"
              style={{
                background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)',
                color: '#FFFFFF',
                boxShadow: '0 6px 20px rgba(255,90,31,0.30)',
              }}
            >
              <Home className="w-5 h-5" />
              Return to Home
            </Link>
            <button
              id="exportResultsBtn"
              className="flex items-center justify-center gap-2 px-6 py-4 rounded-2xl font-bold text-sm transition-all"
              style={{ background: '#F4F2EF', color: '#78716C', border: '1.5px solid #E7E5E4' }}
              title="Export coming with backend integration"
            >
              <Download className="w-4 h-4" />
              Export Results
            </button>
          </motion.div>

        </div>
      </main>
    </AppShell>
  )
}
