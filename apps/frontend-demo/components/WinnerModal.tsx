'use client'

import { motion, AnimatePresence } from 'motion/react'
import { Player } from '@/types/player'
import { BingoPattern } from '@/types/game'
import { Star, X } from 'lucide-react'

interface WinnerModalProps {
  isOpen: boolean
  winner: Player | null
  placement: 1 | 2 | 3
  pattern: BingoPattern | string
  onClose: () => void
}

const PLACEMENT_CONFIG = {
  1: {
    emoji: '🏆',
    label: '1st Place',
    gradient: 'linear-gradient(135deg, #FBBF24 0%, #F59E0B 100%)',
    glow: 'rgba(245, 158, 11, 0.30)',
    accent: '#B45309',
    bg: '#FFFBEB',
    chipBg: '#FEF3C7',
    chipText: '#B45309',
  },
  2: {
    emoji: '🥈',
    label: '2nd Place',
    gradient: 'linear-gradient(135deg, #D6D3D1 0%, #A8A29E 100%)',
    glow: 'rgba(168, 162, 158, 0.25)',
    accent: '#57534E',
    bg: '#FAFAF9',
    chipBg: '#F4F2EF',
    chipText: '#57534E',
  },
  3: {
    emoji: '🥉',
    label: '3rd Place',
    gradient: 'linear-gradient(135deg, #FFA070 0%, #FF7A42 100%)',
    glow: 'rgba(255, 90, 31, 0.25)',
    accent: '#C23208',
    bg: '#FFF4F0',
    chipBg: '#FFE4D9',
    chipText: '#C23208',
  },
}

export function WinnerModal({ isOpen, winner, placement, pattern, onClose }: WinnerModalProps) {
  const config = PLACEMENT_CONFIG[placement] ?? PLACEMENT_CONFIG[1]

  return (
    <AnimatePresence>
      {isOpen && winner && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          style={{ background: 'rgba(28, 25, 23, 0.65)', backdropFilter: 'blur(8px)' }}
          onClick={onClose}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.85, y: 30 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 20 }}
            transition={{ type: 'spring', stiffness: 280, damping: 22 }}
            className="w-full max-w-[440px] rounded-3xl overflow-hidden relative"
            style={{
              background: '#FFFFFF',
              boxShadow: `0 24px 60px ${config.glow}, 0 8px 24px rgba(0,0,0,0.12)`,
            }}
            onClick={e => e.stopPropagation()}
          >
            {/* Close button */}
            <button
              onClick={onClose}
              className="absolute top-4 right-4 w-8 h-8 rounded-xl flex items-center justify-center z-10 transition-all"
              style={{ background: '#F4F2EF', color: '#A8A29E' }}
              aria-label="Close"
            >
              <X className="w-4 h-4" />
            </button>

            {/* Colored top section */}
            <div
              className="px-8 pt-10 pb-8 flex flex-col items-center text-center"
              style={{ background: config.bg }}
            >
              {/* Trophy / medal */}
              <motion.div
                animate={{ y: [0, -8, 0] }}
                transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
                className="text-6xl mb-4 select-none"
              >
                {config.emoji}
              </motion.div>

              {/* BINGO CONFIRMED label */}
              <p
                className="text-xs font-extrabold uppercase tracking-[0.2em] mb-2"
                style={{ color: config.accent }}
              >
                Bingo Confirmed
              </p>

              {/* Winner name */}
              <h2
                className="text-3xl font-black mb-3 leading-tight"
                style={{ color: '#1C1917' }}
              >
                {winner.name}
              </h2>

              {/* Placement chip */}
              <div
                className="px-5 py-1.5 rounded-full text-sm font-extrabold uppercase tracking-wider"
                style={{ background: config.chipBg, color: config.chipText }}
              >
                {config.label}
              </div>
            </div>

            {/* Details */}
            <div className="px-6 py-6">
              <div
                className="flex items-center justify-between px-4 py-3.5 rounded-2xl mb-6"
                style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
              >
                <div>
                  <p className="text-[10px] font-extrabold uppercase tracking-widest mb-0.5" style={{ color: '#A8A29E' }}>
                    Winning Pattern
                  </p>
                  <p className="text-sm font-bold" style={{ color: '#44403C' }}>{pattern}</p>
                </div>
                {/* Mini bingo icon */}
                <div
                  className="w-10 h-10 rounded-xl p-1.5 grid grid-cols-3 gap-0.5 opacity-60"
                  style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}
                >
                  {[1,2,3,4,5,6,7,8,9].map(i => (
                    <div
                      key={i}
                      className="rounded-sm"
                      style={{ background: i >= 4 && i <= 6 ? '#FF5A1F' : '#F0EDE8' }}
                    />
                  ))}
                </div>
              </div>

              <button
                onClick={onClose}
                className="w-full py-4 rounded-2xl font-extrabold text-base transition-all hover:opacity-90 active:scale-97"
                style={{
                  background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 100%)',
                  color: '#FFFFFF',
                  boxShadow: '0 6px 20px rgba(255, 90, 31, 0.35)',
                }}
              >
                Continue Game
              </button>

              <p className="text-center text-[10px] font-semibold mt-4" style={{ color: '#D6D3D1' }}>
                Validation complete · Activity logged
              </p>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
