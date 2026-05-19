'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { motion } from 'motion/react'
import { ArrowRight, Shield, Target, Trophy, Bot } from 'lucide-react'
import { DecorativeBlobs } from '@/components/illustrations/DecorativeBlobs'
import { BingoCharacter } from '@/components/illustrations/BingoCharacter'

export default function LandingPage() {
  const [gameCode, setGameCode] = useState('')
  const [playerName, setPlayerName] = useState('')
  const router = useRouter()

  const handleJoin = () => {
    if (gameCode.trim()) {
      const code = encodeURIComponent(gameCode.trim())
      const name = encodeURIComponent(playerName.trim() || 'Player')
      router.push(`/lobby?code=${code}&name=${name}`)
    }
  }

  return (
    <div
      className="min-h-screen flex flex-col relative overflow-hidden"
      style={{ background: '#FAF8F5', fontFamily: "'Nunito', ui-rounded, system-ui, sans-serif" }}
    >
      {/* Animated decorative blobs */}
      <DecorativeBlobs variant="landing" />

      {/* Nav bar */}
      <nav className="relative z-10 h-16 px-6 sm:px-10 flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <motion.div
            whileHover={{ scale: 1.08, rotate: 3 }}
            className="w-9 h-9 rounded-lg flex items-center justify-center text-white font-black text-lg"
            style={{ background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)', boxShadow: '0 4px 12px rgba(255,90,31,0.30)' }}
          >
            B
          </motion.div>
          <span className="font-black text-lg tracking-tight" style={{ color: '#1C1917' }}>Virtual Bingo</span>
        </div>
        <Link
          href="/host"
          className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-bold transition-all hover:scale-105"
          style={{ background: '#F4F2EF', color: '#44403C' }}
        >
          <Shield className="w-4 h-4" />
          <span className="hidden sm:inline">Host a Game</span>
        </Link>
      </nav>

      {/* Main content */}
      <main className="relative z-10 flex-1 flex items-center justify-center px-4 py-10">
        <div className="w-full max-w-lg">

          {/* Hero section */}
          <motion.div
            initial={{ opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, ease: [0.25, 0.4, 0.25, 1] }}
            className="text-center mb-10"
          >
            {/* Character + Icon together */}
            <div className="flex items-center justify-center gap-4 mb-6">
              {/* Illustrated character */}
              <motion.div
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.3, type: 'spring' }}
              >
                <BingoCharacter mood="excited" size={90} />
              </motion.div>

              {/* Big icon */}
              <motion.div
                animate={{ y: [0, -10, 0] }}
                transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut' }}
                className="inline-flex items-center justify-center w-24 h-24 rounded-2xl"
                style={{
                  background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 50%, #E8440A 100%)',
                  boxShadow: '0 12px 40px rgba(255, 90, 31, 0.35)',
                }}
              >
                <span className="text-5xl font-black text-white select-none">B</span>
              </motion.div>

              {/* Second character on right */}
              <motion.div
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.4, type: 'spring' }}
              >
                <BingoCharacter mood="happy" size={80} />
              </motion.div>
            </div>

            <h1 className="text-5xl sm:text-6xl font-black tracking-tight mb-4" style={{ color: '#1C1917', lineHeight: 1.1 }}>
              Virtual<br />
              <span style={{ color: '#FF5A1F' }}>Bingo</span>
            </h1>
            <p className="text-base sm:text-lg font-semibold max-w-sm mx-auto leading-relaxed" style={{ color: '#78716C' }}>
              Centralized cards, live word calls, and real-time leaderboards for your next team event.
            </p>
          </motion.div>

          {/* Card */}
          <motion.div
            initial={{ opacity: 0, y: 32 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.15, ease: [0.25, 0.4, 0.25, 1] }}
            className="rounded-2xl p-6 sm:p-8 relative overflow-hidden"
            style={{
              background: '#FFFFFF',
              border: '1.5px solid #F0EDE8',
              boxShadow: '0 8px 40px rgba(0,0,0,0.06)',
            }}
          >
            {/* Subtle corner decoration */}
            <div
              className="absolute top-0 right-0 w-24 h-24 pointer-events-none"
              style={{
                background: 'radial-gradient(circle at top right, rgba(124,92,252,0.06) 0%, transparent 70%)',
              }}
            />

            {/* Sign in with Microsoft (placeholder) */}
            <button
              disabled
              className="w-full flex items-center justify-between px-4 py-3.5 rounded-xl text-sm font-bold mb-6 transition-all"
              style={{ background: '#F4F2EF', color: '#A8A29E', cursor: 'not-allowed', border: '1.5px solid #E7E5E4' }}
            >
              <div className="flex items-center gap-3">
                {/* Microsoft icon */}
                <svg width="16" height="16" viewBox="0 0 21 21" fill="none">
                  <rect x="1" y="1" width="9" height="9" fill="#F25022"/>
                  <rect x="11" y="1" width="9" height="9" fill="#7FBA00"/>
                  <rect x="1" y="11" width="9" height="9" fill="#00A4EF"/>
                  <rect x="11" y="11" width="9" height="9" fill="#FFB900"/>
                </svg>
                Sign in with Microsoft Entra
              </div>
              <span
                className="text-[10px] font-extrabold uppercase tracking-widest px-2.5 py-1 rounded-md shrink-0"
                style={{ background: '#E7E5E4', color: '#B8B2AC' }}
              >
                Soon
              </span>
            </button>

            {/* Divider */}
            <div className="relative mb-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full" style={{ borderTop: '1.5px solid #F0EDE8' }} />
              </div>
              <div className="relative flex justify-center">
                <span className="px-4 text-[11px] uppercase font-extrabold tracking-widest" style={{ background: '#FFFFFF', color: '#D6D3D1' }}>
                  Join with a Code
                </span>
              </div>
            </div>

            {/* Form */}
            <div className="space-y-3">
              <motion.div
                initial={{ opacity: 0, x: -12 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.25 }}
              >
                <input
                  id="playerName"
                  type="text"
                  placeholder="Your Name"
                  value={playerName}
                  onChange={e => setPlayerName(e.target.value)}
                  aria-label="Player name"
                  className="w-full px-4 py-3.5 rounded-xl text-sm font-bold transition-all outline-none"
                  style={{
                    background: '#FAF8F5',
                    border: '1.5px solid #E7E5E4',
                    color: '#1C1917',
                    fontFamily: 'inherit',
                  }}
                  onFocus={e => { e.target.style.border = '1.5px solid #FF5A1F'; e.target.style.background = '#FFFFFF'; e.target.style.boxShadow = '0 0 0 4px rgba(255,90,31,0.08)'; }}
                  onBlur={e => { e.target.style.border = '1.5px solid #E7E5E4'; e.target.style.background = '#FAF8F5'; e.target.style.boxShadow = 'none'; }}
                />
              </motion.div>
              <motion.div
                initial={{ opacity: 0, x: -12 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.3 }}
                className="flex gap-3"
              >
                <input
                  id="gameCode"
                  type="text"
                  placeholder="GAME CODE"
                  value={gameCode}
                  onChange={e => setGameCode(e.target.value.toUpperCase())}
                  onKeyDown={e => e.key === 'Enter' && handleJoin()}
                  aria-label="Game code"
                  className="flex-1 px-4 py-3.5 rounded-xl text-sm font-black uppercase tracking-widest text-center transition-all outline-none"
                  style={{
                    background: '#FAF8F5',
                    border: '1.5px solid #E7E5E4',
                    color: '#1C1917',
                    letterSpacing: '0.15em',
                    fontFamily: 'inherit',
                  }}
                  onFocus={e => { e.target.style.border = '1.5px solid #FF5A1F'; e.target.style.background = '#FFFFFF'; e.target.style.boxShadow = '0 0 0 4px rgba(255,90,31,0.08)'; }}
                  onBlur={e => { e.target.style.border = '1.5px solid #E7E5E4'; e.target.style.background = '#FAF8F5'; e.target.style.boxShadow = 'none'; }}
                />
                <motion.button
                  whileTap={{ scale: 0.95 }}
                  whileHover={{ scale: 1.05, y: -2 }}
                  onClick={handleJoin}
                  id="joinGameBtn"
                  aria-label="Join game"
                  className="flex items-center justify-center gap-2 px-6 rounded-xl font-extrabold text-sm transition-all"
                  style={{
                    background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)',
                    color: '#FFFFFF',
                    boxShadow: '0 4px 16px rgba(255,90,31,0.35)',
                    minWidth: '90px',
                  }}
                >
                  Join
                  <ArrowRight className="w-4 h-4" />
                </motion.button>
              </motion.div>
            </div>
          </motion.div>

          {/* Features strip */}
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
            className="mt-6 grid grid-cols-3 gap-3"
          >
            {[
              { icon: <Target className="w-6 h-6" style={{ color: '#FF5A1F' }} />, label: 'Live Cards', desc: 'Auto-generated' },
              { icon: <Trophy className="w-6 h-6" style={{ color: '#F59E0B' }} />, label: 'Leaderboard', desc: 'Real-time' },
              { icon: <Bot className="w-6 h-6" style={{ color: '#7C5CFC' }} />, label: 'AI Host', desc: 'Commentary' },
            ].map((f, i) => (
              <motion.div
                key={f.label}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.45 + i * 0.08 }}
                className="rounded-xl p-3 text-center flex flex-col items-center"
                style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}
              >
                <div className="mb-2 flex items-center justify-center">{f.icon}</div>
                <p className="text-[10px] font-extrabold" style={{ color: '#1C1917' }}>{f.label}</p>
                <p className="text-[9px] font-semibold" style={{ color: '#A8A29E' }}>{f.desc}</p>
              </motion.div>
            ))}
          </motion.div>

          {/* Host & Admin links */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
            className="mt-6 flex flex-col gap-2"
          >
            <Link
              href="/host"
              id="hostGameLink"
              className="w-full flex items-center justify-between gap-3 px-4 py-3.5 rounded-xl text-sm font-bold transition-all hover:brightness-95 active:scale-[0.99]"
              style={{
                background: '#F4F2EF',
                color: '#57534E',
                border: '1.5px solid #E7E5E4',
              }}
            >
              <div className="flex items-center gap-3">
                <div className="w-7 h-7 rounded-lg flex items-center justify-center shrink-0" style={{ background: '#E7E5E4' }}>
                  <Shield className="w-3.5 h-3.5" style={{ color: '#78716C' }} />
                </div>
                I&apos;m a host, take me to the dashboard
              </div>
              <ArrowRight className="w-3.5 h-3.5 shrink-0" style={{ color: '#A8A29E' }} />
            </Link>
            <Link
              href="/admin"
              id="adminDashboardLink"
              className="w-full flex items-center justify-between gap-3 px-4 py-3.5 rounded-xl text-sm font-bold transition-all hover:brightness-95 active:scale-[0.99]"
              style={{
                background: '#FAFAF9',
                color: '#A8A29E',
                border: '1.5px solid #F0EDE8',
              }}
            >
              <div className="flex items-center gap-3">
                <div className="w-7 h-7 rounded-lg flex items-center justify-center shrink-0" style={{ background: '#F0EDE8' }}>
                  <Shield className="w-3.5 h-3.5" style={{ color: '#A8A29E' }} />
                </div>
                Admin Operations Center
              </div>
              <ArrowRight className="w-3.5 h-3.5 shrink-0" style={{ color: '#C4BFB9' }} />
            </Link>
          </motion.div>

          {/* Footer */}
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.55 }}
            className="text-center text-xs font-semibold mt-8"
            style={{ color: '#D6D3D1' }}
          >
            Internal corporate tool · Authentication coming soon
          </motion.p>
        </div>
      </main>
    </div>
  )
}
