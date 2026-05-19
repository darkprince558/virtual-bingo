'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { motion } from 'motion/react'
import { ArrowRight, Shield, Sparkles } from 'lucide-react'

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
      {/* Decorative background blobs */}
      <div
        className="absolute top-[-180px] right-[-180px] w-[500px] h-[500px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(255,164,112,0.20) 0%, transparent 70%)' }}
      />
      <div
        className="absolute bottom-[-150px] left-[-150px] w-[450px] h-[450px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(124,92,252,0.12) 0%, transparent 70%)' }}
      />
      <div
        className="absolute top-[40%] left-[10%] w-[200px] h-[200px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(34,170,106,0.08) 0%, transparent 70%)' }}
      />

      {/* Nav bar */}
      <nav className="relative z-10 h-16 px-6 sm:px-10 flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <div
            className="w-9 h-9 rounded-2xl flex items-center justify-center text-white font-black text-lg"
            style={{ background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)', boxShadow: '0 4px 12px rgba(255,90,31,0.30)' }}
          >
            B
          </div>
          <span className="font-black text-lg tracking-tight" style={{ color: '#1C1917' }}>Virtual Bingo</span>
        </div>
        <Link
          href="/host"
          className="flex items-center gap-2 px-4 py-2 rounded-full text-sm font-bold transition-all"
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
            {/* Big icon */}
            <motion.div
              animate={{ y: [0, -10, 0] }}
              transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut' }}
              className="inline-flex items-center justify-center w-24 h-24 rounded-3xl mb-6"
              style={{
                background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 50%, #E8440A 100%)',
                boxShadow: '0 12px 40px rgba(255, 90, 31, 0.35)',
              }}
            >
              <span className="text-5xl font-black text-white select-none">B</span>
            </motion.div>

            <h1 className="text-5xl sm:text-6xl font-black tracking-tight mb-4" style={{ color: '#1C1917', lineHeight: 1.1 }}>
              Virtual<br />
              <span style={{ color: '#FF5A1F' }}>Bingo</span>
            </h1>
            <p className="text-base sm:text-lg font-semibold max-w-sm mx-auto leading-relaxed" style={{ color: '#78716C' }}>
              Centralized cards, live word calls, and real-time leaderboards — for your next team event.
            </p>
          </motion.div>

          {/* Card */}
          <motion.div
            initial={{ opacity: 0, y: 32 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.15, ease: [0.25, 0.4, 0.25, 1] }}
            className="rounded-3xl p-6 sm:p-8"
            style={{
              background: '#FFFFFF',
              border: '1.5px solid #F0EDE8',
              boxShadow: '0 8px 40px rgba(0,0,0,0.06)',
            }}
          >
            {/* Sign in with Microsoft (placeholder) */}
            <button
              disabled
              className="w-full flex items-center justify-center gap-3 py-3.5 rounded-2xl text-sm font-bold mb-6 transition-all"
              style={{ background: '#F4F2EF', color: '#A8A29E', cursor: 'not-allowed', border: '1.5px solid #E7E5E4' }}
            >
              {/* Microsoft icon */}
              <svg width="18" height="18" viewBox="0 0 21 21" fill="none">
                <rect x="1" y="1" width="9" height="9" fill="#F25022"/>
                <rect x="11" y="1" width="9" height="9" fill="#7FBA00"/>
                <rect x="1" y="11" width="9" height="9" fill="#00A4EF"/>
                <rect x="11" y="11" width="9" height="9" fill="#FFB900"/>
              </svg>
              Sign in with Microsoft Entra
              <span className="ml-auto text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-full" style={{ background: '#E7E5E4', color: '#A8A29E' }}>Soon</span>
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
              <input
                id="playerName"
                type="text"
                placeholder="Your Name"
                value={playerName}
                onChange={e => setPlayerName(e.target.value)}
                aria-label="Player name"
                className="w-full px-4 py-3.5 rounded-2xl text-sm font-bold transition-all outline-none"
                style={{
                  background: '#FAF8F5',
                  border: '1.5px solid #E7E5E4',
                  color: '#1C1917',
                  fontFamily: 'inherit',
                }}
                onFocus={e => { e.target.style.border = '1.5px solid #FF5A1F'; e.target.style.background = '#FFFFFF'; }}
                onBlur={e => { e.target.style.border = '1.5px solid #E7E5E4'; e.target.style.background = '#FAF8F5'; }}
              />
              <div className="flex gap-3">
                <input
                  id="gameCode"
                  type="text"
                  placeholder="GAME CODE"
                  value={gameCode}
                  onChange={e => setGameCode(e.target.value.toUpperCase())}
                  onKeyDown={e => e.key === 'Enter' && handleJoin()}
                  aria-label="Game code"
                  className="flex-1 px-4 py-3.5 rounded-2xl text-sm font-black uppercase tracking-widest text-center transition-all outline-none"
                  style={{
                    background: '#FAF8F5',
                    border: '1.5px solid #E7E5E4',
                    color: '#1C1917',
                    letterSpacing: '0.15em',
                    fontFamily: 'inherit',
                  }}
                  onFocus={e => { e.target.style.border = '1.5px solid #FF5A1F'; e.target.style.background = '#FFFFFF'; }}
                  onBlur={e => { e.target.style.border = '1.5px solid #E7E5E4'; e.target.style.background = '#FAF8F5'; }}
                />
                <motion.button
                  whileTap={{ scale: 0.95 }}
                  whileHover={{ scale: 1.03 }}
                  onClick={handleJoin}
                  id="joinGameBtn"
                  aria-label="Join game"
                  className="flex items-center justify-center gap-2 px-6 rounded-2xl font-extrabold text-sm transition-all"
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
              </div>
            </div>
          </motion.div>

          {/* Host & Admin links */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 }}
            className="text-center mt-6 flex flex-col items-center gap-2"
          >
            <Link
              href="/host"
              id="hostGameLink"
              className="inline-flex items-center gap-2 text-sm font-bold px-5 py-3 rounded-full transition-all"
              style={{ color: '#A8A29E' }}
            >
              <Shield className="w-4 h-4" />
              I&apos;m a host — take me to the dashboard
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
            <Link
              href="/admin"
              id="adminDashboardLink"
              className="inline-flex items-center gap-2 text-sm font-bold px-5 py-3 rounded-full transition-all"
              style={{ color: '#D6D3D1' }}
            >
              <Shield className="w-4 h-4" />
              Admin Operations Center
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </motion.div>

          {/* Footer */}
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
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
