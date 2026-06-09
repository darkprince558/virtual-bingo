'use client'

import { motion } from 'motion/react'
import { LeaderboardEntry } from '@/types/player'

interface LeaderboardProps {
  entries: LeaderboardEntry[]
}

/** Custom SVG trophy icons instead of emoji - warmer, Headspace-style */
function TrophyIcon({ placement }: { placement: number }) {
  if (placement === 1) {
    return (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
        <path d="M8 21h8M12 17v4M6 3h12v4a6 6 0 01-12 0V3z" stroke="#FBBF24" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M6 3H3v3a3 3 0 003 3M18 3h3v3a3 3 0 01-3 3" stroke="#F59E0B" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <circle cx="12" cy="8" r="2" fill="#FBBF24" />
      </svg>
    )
  }
  if (placement === 2) {
    return (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="10" r="7" stroke="#A8A29E" strokeWidth="2" />
        <path d="M12 6v4l2 2" stroke="#A8A29E" strokeWidth="2" strokeLinecap="round" />
        <path d="M12 17v4M9 21h6" stroke="#D6D3D1" strokeWidth="2" strokeLinecap="round" />
      </svg>
    )
  }
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
      <path d="M12 2l2.4 7.2H22l-6 4.8 2.4 7.2L12 16.4 5.6 21.2 8 14 2 9.2h7.6L12 2z" stroke="#FFA070" strokeWidth="2" fill="none" strokeLinejoin="round" />
    </svg>
  )
}

const PLACE_STYLES: Record<number, { bg: string; text: string; border: string; glowColor: string }> = {
  1: { bg: '#FEF3C7', text: '#B45309', border: '#FCD34D', glowColor: 'rgba(251,191,36,0.3)' },
  2: { bg: '#F4F2EF', text: '#57534E', border: '#D6D3D1', glowColor: 'rgba(168,162,158,0.15)' },
  3: { bg: '#FFF0F3', text: '#C23208', border: '#FFB0C0', glowColor: 'rgba(232,0,45,0.15)' },
}

function getInitials(name: string) {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

export function Leaderboard({ entries }: LeaderboardProps) {
  return (
    <div className="flex flex-col">
      <div className="flex items-center gap-2 mb-4">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
          <path d="M8 21h8M12 17v4M6 3h12v4a6 6 0 01-12 0V3z" stroke="#F59E0B" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
        <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
          Leaderboard
        </span>
      </div>

      <div className="space-y-2.5">
        {entries.map((entry, i) => {
          const style = PLACE_STYLES[entry.placement] ?? { bg: '#FAFAF9', text: '#78716C', border: 'transparent', glowColor: 'transparent' }
          return (
            <motion.div
              key={entry.player.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: i * 0.1, type: 'spring', stiffness: 300, damping: 25 }}
              className="flex items-center gap-3 px-3.5 py-3 rounded-md transition-all relative overflow-hidden"
              style={{
                background: style.bg,
                border: `1.5px solid ${style.border}`,
                boxShadow: entry.placement === 1 ? `0 4px 16px ${style.glowColor}` : 'none',
              }}
            >
              {/* Golden shimmer for 1st place */}
              {entry.placement === 1 && (
                <div
                  className="absolute inset-0 animate-golden pointer-events-none"
                  style={{
                    background: 'linear-gradient(90deg, transparent 0%, rgba(251,191,36,0.08) 50%, transparent 100%)',
                    backgroundSize: '200% 100%',
                  }}
                />
              )}

              {/* Placement icon */}
              <div className="shrink-0 w-7 flex items-center justify-center relative z-10">
                <TrophyIcon placement={entry.placement} />
              </div>

              {/* Avatar */}
              <div
                className="w-8 h-8 rounded-md flex items-center justify-center text-xs font-black shrink-0 relative z-10"
                style={{
                  background: entry.placement === 1 ? 'linear-gradient(135deg, #FBBF24, #F59E0B)' : '#FFFFFF',
                  color: entry.placement === 1 ? '#FFFFFF' : '#78716C',
                  border: `1.5px solid ${style.border}`,
                }}
              >
                {getInitials(entry.player.name)}
              </div>

              {/* Name & match count */}
              <div className="flex-1 min-w-0 relative z-10">
                <p className="text-sm font-bold truncate" style={{ color: style.text }}>
                  {entry.player.name}
                </p>
                <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>
                  {entry.wordsMatched} matched
                </p>
              </div>
            </motion.div>
          )
        })}

        {entries.length === 0 && (
          <div className="text-center py-6">
            <motion.div
              animate={{ y: [0, -4, 0] }}
              transition={{ duration: 3, repeat: Infinity }}
              className="inline-block mb-2"
            >
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none">
                <path d="M8 21h8M12 17v4M6 3h12v4a6 6 0 01-12 0V3z" stroke="#E7E5E4" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </motion.div>
            <p className="text-sm font-semibold" style={{ color: '#D6D3D1' }}>
              No scores yet
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
