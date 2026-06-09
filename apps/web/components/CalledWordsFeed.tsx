'use client'

import { motion, AnimatePresence } from 'motion/react'
import { CalledWord } from '@/types/game'

interface CalledWordsFeedProps {
  words: CalledWord[]
}

function timeAgo(isoString: string): string {
  const diff = Math.floor((Date.now() - new Date(isoString).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

export function CalledWordsFeed({ words }: CalledWordsFeedProps) {
  const reversed = [...words].reverse()

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 mb-3">
        {/* Custom clock icon */}
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6.5" stroke="#A8A29E" strokeWidth="1.5" />
          <path d="M8 4.5v4l2.5 1.5" stroke="#A8A29E" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
        <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
          Called Words
        </span>
        <span
          className="ml-auto text-xs font-bold px-2 py-0.5 rounded-full"
          style={{ background: '#F4F2EF', color: '#78716C' }}
        >
          {words.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto space-y-2 pr-1">
        {reversed.length === 0 ? (
          <div className="text-center py-8">
            <motion.div
              animate={{ opacity: [0.3, 0.7, 0.3] }}
              transition={{ duration: 2, repeat: Infinity }}
              className="mx-auto mb-2"
            >
              <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
                <rect x="4" y="8" width="24" height="16" rx="4" stroke="#E7E5E4" strokeWidth="2" />
                <line x1="10" y1="14" x2="22" y2="14" stroke="#E7E5E4" strokeWidth="2" strokeLinecap="round" />
                <line x1="10" y1="18" x2="18" y2="18" stroke="#E7E5E4" strokeWidth="2" strokeLinecap="round" />
              </svg>
            </motion.div>
            <p className="text-sm font-semibold" style={{ color: '#D6D3D1' }}>No words called yet</p>
          </div>
        ) : (
          <AnimatePresence initial={false}>
            {reversed.map((w, i) => (
              <motion.div
                key={w.id}
                initial={{ opacity: 0, x: -16, scale: 0.95 }}
                animate={{ opacity: 1, x: 0, scale: 1 }}
                transition={{ type: 'spring', stiffness: 300, damping: 25 }}
                className="flex items-center justify-between px-3.5 py-2.5 rounded-md transition-all relative overflow-hidden"
                style={{
                  background: i === 0 ? '#FFF0F3' : '#FAFAF9',
                  border: i === 0 ? '1.5px solid #FFE4D9' : '1.5px solid transparent',
                }}
              >
                {/* Pulse glow on latest word */}
                {i === 0 && (
                  <motion.div
                    className="absolute inset-0 pointer-events-none"
                    animate={{ opacity: [0.1, 0.3, 0.1] }}
                    transition={{ duration: 2, repeat: Infinity }}
                    style={{ background: 'linear-gradient(135deg, rgba(232,0,45,0.08), transparent)' }}
                  />
                )}

                <div className="flex items-center gap-2.5 relative z-10">
                  {/* Call number */}
                  <span
                    className="text-[9px] font-black w-5 text-center shrink-0"
                    style={{ color: i === 0 ? '#E8002D' : '#D6D3D1' }}
                  >
                    #{reversed.length - i}
                  </span>
                  <span
                    className="text-sm font-bold leading-tight"
                    style={{ color: i === 0 ? '#C40026' : '#44403C' }}
                  >
                    {w.word}
                  </span>
                </div>
                <span className="text-[10px] font-semibold shrink-0 ml-3 relative z-10" style={{ color: '#A8A29E' }}>
                  {timeAgo(w.calledAt)}
                </span>
              </motion.div>
            ))}
          </AnimatePresence>
        )}
      </div>
    </div>
  )
}
